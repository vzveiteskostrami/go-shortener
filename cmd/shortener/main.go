// Сервер сокращения URL. Принимает полный URL на входе, возвращает сокращённый.
// При обращении по сокращённому URL делает переадресацию на полный URL. Ведение
// базы данных URL. Поддерживается владелец и действия по вводу новых URL и удаление
// ненужных.
// Запуск в командной строке:
//
//	shortener [-a=<[in host]:<in port>>] [-b=<[out host]:<out port>>] [-f=<Storage text file name>] [-d=<Database connect string>]
package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/vzveiteskostrami/go-shortener/internal/auth"
	"github.com/vzveiteskostrami/go-shortener/internal/compressing"
	"github.com/vzveiteskostrami/go-shortener/internal/config"
	"github.com/vzveiteskostrami/go-shortener/internal/dbf"
	"github.com/vzveiteskostrami/go-shortener/internal/logging"
	"github.com/vzveiteskostrami/go-shortener/internal/shorturl"

	"github.com/go-chi/chi/v5"

	"net/http/pprof"
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
	fmt.Println("Build date:", buildCommit)

	logging.LoggingInit()
	defer logging.LoggingSync()
	config.ReadData()
	dbf.MakeStorage()
	shorturl.SetURLNum(dbf.Store.DBFInit())
	defer dbf.Store.DBFClose()
	//shorturl.GoDel()
	go shorturl.DoDel()

	srv = &http.Server{
		Addr:        config.Addresses.In.Host + ":" + strconv.Itoa(config.Addresses.In.Port),
		Handler:     mainRouter(),
		IdleTimeout: time.Second * 1,
	}

	logging.S().Infow(
		"Starting server",
		"addr", config.Addresses.In.Host+":"+strconv.Itoa(config.Addresses.In.Port),
	)

	logging.S().Fatal(srv.ListenAndServe())
}

func mainRouter() chi.Router {
	r := chi.NewRouter()

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
