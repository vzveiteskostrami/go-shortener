// Сервер сокращения URL. Принимает полный URL на входе, возвращает сокращённый.
// При обращении по сокращённому URL делает переадресацию на полный URL. Ведение
// базы данных URL. Поддерживается владелец и действия по вводу новых URL и удаление
// ненужных.
// Запуск в командной строке:
//
//	shortener [-a=<[in host]:<in port>>] [-b=<[out host]:<out port>>] [-f=<Storage text file name>] [-d=<Database connect string>]
package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/vzveiteskostrami/go-shortener/internal/auth"
	"github.com/vzveiteskostrami/go-shortener/internal/compressing"
	"github.com/vzveiteskostrami/go-shortener/internal/config"
	"github.com/vzveiteskostrami/go-shortener/internal/dbf"
	"github.com/vzveiteskostrami/go-shortener/internal/logging"
	"github.com/vzveiteskostrami/go-shortener/internal/shorturl"
	trust "github.com/vzveiteskostrami/go-shortener/internal/trusted"
	"github.com/vzveiteskostrami/go-shortener/internal/urlman"
	"golang.org/x/crypto/acme/autocert"

	"github.com/go-chi/chi/v5"

	"github.com/vzveiteskostrami/go-shortener/internal/lgrpc"
)

var (
	srv          *http.Server
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}
	fmt.Println("Build version:", buildVersion)
	fmt.Println("Build date:", buildDate)
	fmt.Println("Build commit:", buildCommit)

	logging.LoggingInit()
	config.ReadData()
	dbf.MakeStorage()
	urlman.SetURLNum(dbf.Store.DBFInit())
	defer dbf.Store.DBFClose()
	go shorturl.DoDel()

	if config.UsegRPC {
		lgrpc.DogRPC()
	} else {
		doHTTP()
	}
}

func mainRouter() chi.Router {
	r := chi.NewRouter()

	r.Route("/api/internal", func(r chi.Router) {
		r.Use(compressing.GZIPHandle)
		r.Use(logging.WithLogging)
		r.Use(trust.TrustedHandle)
		r.Use(auth.AuthHandle)
		r.Get("/stats", shorturl.GetStatsf)
	})

	r.Route("/api", func(r chi.Router) {
		r.Use(compressing.GZIPHandle)
		r.Use(logging.WithLogging)
		r.Use(auth.AuthHandle)
		r.Post("/shorten", shorturl.SetJSONLinkf)
		r.Post("/shorten/batch", shorturl.SetJSONBatchLinkf)
		r.Get("/user/urls", shorturl.GetOwnerURLsListf)
		r.Delete("/user/urls", shorturl.DeleteOwnerURLsListf)
	})

	r.Route("/ping", func(r chi.Router) {
		r.Use(logging.WithLogging)
		r.Get("/", dbf.Store.PingDBf)
	})

	r.Route("/{shlink}", func(r chi.Router) {
		r.Use(compressing.GZIPHandle)
		r.Use(logging.WithLogging)
		r.Get("/", shorturl.GetLinkf)
	})

	r.Route("/", func(r chi.Router) {
		r.Use(compressing.GZIPHandle)
		r.Use(logging.WithLogging)
		r.Use(auth.AuthHandle)
		r.Post("/", shorturl.SetLinkf)
	})

	addPprof(r)

	return r
}

func addPprof(r *chi.Mux) {
	r.HandleFunc("/pprof/*", pprof.Index)
	r.HandleFunc("/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/pprof/profile", pprof.Profile)
	r.HandleFunc("/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/pprof/trace", pprof.Trace)

	r.Handle("/pprof/goroutine", pprof.Handler("goroutine"))
	r.Handle("/pprof/threadcreate", pprof.Handler("threadcreate"))
	r.Handle("/pprof/mutex", pprof.Handler("mutex"))
	r.Handle("/pprof/heap", pprof.Handler("heap"))
	r.Handle("/pprof/block", pprof.Handler("block"))
	r.Handle("/pprof/allocs", pprof.Handler("allocs"))
}

func doHTTP() {
	srv = &http.Server{
		Addr:        config.Addresses.In.Host + ":" + strconv.Itoa(config.Addresses.In.Port),
		Handler:     mainRouter(),
		IdleTimeout: time.Second * 1,
	}

	//idleConnsClosed := make(chan struct{})
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)
	go func() {
		<-sigs
		if err := srv.Shutdown(context.Background()); err != nil {
			logging.S().Errorln("Server shutdown error", err)
		} else {
			logging.S().Infoln("Server has been closed succesfully")
		}
		//close(idleConnsClosed)
	}()

	if config.UseHTTPS {
		logging.S().Infow(
			"Starting server with SSL/TLS",
			"addr", config.Addresses.In.Host+":"+strconv.Itoa(config.Addresses.In.Port),
		)
		manager := &autocert.Manager{
			// директория для хранения сертификатов
			Cache: autocert.DirCache("cache-dir"),
			// функция, принимающая Terms of Service издателя сертификатов
			Prompt: autocert.AcceptTOS,
			// перечень доменов, для которых будут поддерживаться сертификаты
			HostPolicy: autocert.HostWhitelist(config.Addresses.In.Host, "127.0.0.1", "localhost"),
		}
		srv.TLSConfig = manager.TLSConfig()
		if err := srv.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
			logging.S().Fatal(err)
		}
	} else {
		logging.S().Infow(
			"Starting server",
			"addr", config.Addresses.In.Host+":"+strconv.Itoa(config.Addresses.In.Port),
		)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			logging.S().Fatal(err)
		}
	}
	//<-idleConnsClosed
	logging.S().Infoln("Major thread go home")
}
