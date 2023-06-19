package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	config "github.com/vzveiteskostrami/go-shortener/cmd/shortener/Cfg"

	"github.com/go-chi/chi/v5"
)

var (
	currURLNum  int64
	store       map[string]string
	lockCounter sync.Mutex
	srv         *http.Server
)

func main() {
	config.ReadData()
	currURLNum = 0
	store = make(map[string]string)

	srv = &http.Server{
		Addr:        config.Addresses.In.Host + ":" + strconv.Itoa(config.Addresses.In.Port),
		Handler:     mainRouter(),
		IdleTimeout: time.Second * 1,
	}

	fmt.Println("Сервер запущен на " + config.Addresses.In.Host + ":" + strconv.Itoa(config.Addresses.In.Port))

	log.Fatal(srv.ListenAndServe())
}

func closeServer() {
	time.Sleep(100 * time.Millisecond)
	srv.Close()
}

func mainRouter() chi.Router {
	r := chi.NewRouter()

	r.Get("/{shlink}", getLink)
	r.Post("/", setLink)
	return r
}

func getLink(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	link := chi.URLParam(r, "shlink")

	url := store[link]
	if url == "" {
		http.Error(w, `Не найден shortURL `+link, http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func setLink(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	url := string(b)
	if url == "" {
		http.Error(w, `Не указан URL`, http.StatusBadRequest)
		return
	}

	lockCounter.Lock()
	if url == "close" {
		defer lockCounter.Unlock()
		w.Write([]byte("Сервер выключен"))
		go closeServer()
	} else {
		surl := strconv.FormatInt(currURLNum, 36)
		currURLNum++
		lockCounter.Unlock()
		if store == nil {
			store = make(map[string]string)
		}
		store[surl] = url
		w.WriteHeader(http.StatusCreated)

		if config.Addresses.In == nil {
			config.ReadData()
		}

		w.Write([]byte(config.Addresses.Out.Host + ":" + strconv.Itoa(config.Addresses.Out.Port) + "/" + surl))
	}
}
