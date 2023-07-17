package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/vzveiteskostrami/go-shortener/internal/compressing"
	"github.com/vzveiteskostrami/go-shortener/internal/config"
	"github.com/vzveiteskostrami/go-shortener/internal/dbf"
	"github.com/vzveiteskostrami/go-shortener/internal/logging"
	"github.com/vzveiteskostrami/go-shortener/internal/shorturl"

	"github.com/go-chi/chi/v5"
)

var (
	srv *http.Server
)

func main() {
	logging.LoggingInit()
	defer logging.LoggingSync()
	config.ReadData()
	shorturl.SetURLNum(dbf.DBFInit())
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
	r.Use(compressing.GZIPHandle)
	r.Use(logging.WithLogging)

	r.Post("/", shorturl.SetLinkf)
	r.Get("/{shlink}", shorturl.GetLinkf)
	r.Post("/api/shorten", shorturl.SetJSONLinkf)
	r.Post("/api/shorten/batch", shorturl.SetJSONBatchLinkf)
	r.Get("/ping", dbf.PingDBf)

	return r
}
