package surl

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/vzveiteskostrami/go-shortener/internal/config"
	"github.com/vzveiteskostrami/go-shortener/internal/dbf"
	"github.com/vzveiteskostrami/go-shortener/internal/misc"
)

var (
	currURLNum  int64
	lockCounter sync.Mutex
)

func GetLink() http.Handler {
	fn := GetLinkf
	return http.HandlerFunc(fn)
}

func SetURLNum(num int64) {
	currURLNum = num
}

func GetLinkf(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	link := chi.URLParam(r, "shlink")

	url, ok := dbf.FindLink(link)
	if !ok {
		http.Error(w, `Не найден shortURL `+link, http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", url.OriginalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func SetLink() http.Handler {
	fn := SetLinkf
	return http.HandlerFunc(fn)
}

func SetLinkf(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
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
	currNum := currURLNum
	currURLNum++
	lockCounter.Unlock()
	dbf.DBFSaveLink(dbf.StorageURL{OriginalURL: url,
		UUID:     currNum,
		ShortURL: strconv.FormatInt(currNum, 36)})
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(config.MakeURL(currNum)))
}

type inURL struct {
	URL string `json:"url"`
}

type outURL struct {
	Result string `json:"result"`
}

func SetJSONLink() http.Handler {
	fn := SetJSONLinkf
	return http.HandlerFunc(fn)
}

func SetJSONLinkf(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var url inURL
	if err := json.NewDecoder(r.Body).Decode(&url); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if misc.IsNil(url.URL) || url.URL == "" {
		http.Error(w, `Не указан URL`, http.StatusBadRequest)
		return
	}

	var surl outURL
	lockCounter.Lock()
	currNum := currURLNum
	currURLNum++
	lockCounter.Unlock()
	dbf.DBFSaveLink(dbf.StorageURL{UUID: currNum,
		OriginalURL: url.URL,
		ShortURL:    strconv.FormatInt(currNum, 36)})
	w.WriteHeader(http.StatusCreated)
	var buf bytes.Buffer
	surl.Result = config.MakeURL(currNum)

	jsonEncoder := json.NewEncoder(&buf)
	jsonEncoder.Encode(surl)
	w.Write(buf.Bytes())
}
