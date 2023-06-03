package main

import (
	"io"
	"net/http"
	"os"
	"strconv"
)

// -------------------------------------------------------------------------------------
var (
	currURLnum int64
	store      map[string]string
)

// *************************************************************************************
func main() {
	currURLnum = 0
	store = make(map[string]string)

	mux := http.NewServeMux()
	mux.HandleFunc(`/`, entryPoint)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
	os.Exit(0)
}

// *************************************************************************************
func entryPoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	if r.Method == http.MethodPost {
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

		surl := strconv.FormatInt(currURLnum, 36)
		currURLnum++
		store[surl] = url
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("http://localhost:8080/" + surl))
	} else if r.Method == http.MethodGet {
		if len(r.RequestURI) < 2 {
			http.Error(w, "Нераспознанный формат запроса", http.StatusBadRequest)
			return
		}

		url := store[r.RequestURI[1:]]
		if url == "" {
			http.Error(w, `Не найден shortURL `+r.RequestURI[1:], http.StatusBadRequest)
			return
		}

		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		http.Error(w, "Ожидался POST или GET", http.StatusBadRequest)
	}
}
