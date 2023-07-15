package shorturl

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
	currURLNum  int64 = 0
	lockCounter sync.Mutex
)

func GetLink() http.Handler {
	return http.HandlerFunc(GetLinkf)
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
	return http.HandlerFunc(SetLinkf)
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
	w.Write([]byte(makeURL(currNum)))
}

type inURL struct {
	URL string `json:"url"`
}

type outURL struct {
	Result string `json:"result"`
}

func SetJSONLink() http.Handler {
	return http.HandlerFunc(SetJSONLinkf)
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
	surl.Result = makeURL(currNum)

	jsonEncoder := json.NewEncoder(&buf)
	jsonEncoder.Encode(surl)
	w.Write(buf.Bytes())
}

func makeURL(num int64) string {
	if config.Addresses.In == nil {
		config.ReadData()
	}
	return config.Addresses.Out.Host + ":" + strconv.Itoa(config.Addresses.Out.Port) + "/" + strconv.FormatInt(num, 36)
}
