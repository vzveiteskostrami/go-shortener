package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
)

// -------------------------------------------------------------------------------------
var (
	currURLnum  int64
	store       map[string]string
	lockCounter sync.Mutex
	srv         *http.Server
)

// *************************************************************************************
func main() {
	flag.Parse()
	configStart()
	currURLnum = 0
	store = make(map[string]string)

	fmt.Println("Сервер запущен на " + cfg.InAddr.Host + ":" + strconv.Itoa(cfg.InAddr.Port))

	//log.Fatal(http.ListenAndServe(cfg.InAddr.Host+":"+strconv.Itoa(cfg.InAddr.Port), mainRouter()))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(cfg.InAddr.Port), mainRouter()))
}

// *************************************************************************************
func closeServer() {
	time.Sleep(100 * time.Millisecond)
	srv.Close()
}

// *************************************************************************************
func mainRouter() chi.Router {
	r := chi.NewRouter()

	r.Get("/{shlink}", getLink)
	r.Post("/", setLink)
	return r
}

// *************************************************************************************
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

// *************************************************************************************
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
		surl := strconv.FormatInt(currURLnum, 36)
		currURLnum++
		lockCounter.Unlock()
		if store == nil {
			store = make(map[string]string)
		}
		store[surl] = url
		w.WriteHeader(http.StatusCreated)

		if cfg.InAddr == nil {
			configStart()
		}

		w.Write([]byte(cfg.OutAddr.Host + ":" + strconv.Itoa(cfg.OutAddr.Port) + "/" + surl))
	}
}
