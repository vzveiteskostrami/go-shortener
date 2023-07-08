package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/vzveiteskostrami/go-shortener/internal/config"
	"github.com/vzveiteskostrami/go-shortener/internal/cpress"
	"github.com/vzveiteskostrami/go-shortener/internal/dbf"
	"github.com/vzveiteskostrami/go-shortener/internal/logging"
	"github.com/vzveiteskostrami/go-shortener/internal/surl"

	"github.com/go-chi/chi/v5"
)

var (
	srv *http.Server
)

func main() {
	logging.LoggingInit()
	defer logging.LoggingSync()
	config.ReadData()
	surl.SetURLNum(dbf.DBFInit())
	defer dbf.DBFClose()

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
	r.Use(cpress.GZIPHandle)
	r.Use(logging.WithLogging)

	r.Post("/", surl.SetLinkf)
	r.Get("/{shlink}", surl.GetLinkf)
	r.Post("/api/shorten", surl.SetJSONLinkf)

	//r.Handle(`/`, cpress.GZIPHandle(logging.WithLogging(setLink())))
	//r.Handle("/{shlink}", cpress.GZIPHandle(logging.WithLogging(getLink())))
	//r.Handle("/api/shorten", cpress.GZIPHandle(logging.WithLogging(setJSONLink())))
	return r
}
