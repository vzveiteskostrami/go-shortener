package shorturl

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/vzveiteskostrami/go-shortener/internal/auth"
	"github.com/vzveiteskostrami/go-shortener/internal/config"
	"github.com/vzveiteskostrami/go-shortener/internal/dbf"
	"github.com/vzveiteskostrami/go-shortener/internal/logging"
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

	completed := make(chan struct{})
	url := dbf.StorageURL{}
	ok := false

	go func() {
		url, ok = dbf.Store.FindLink(link, true)
		completed <- struct{}{}
	}()

	select {
	case <-completed:
		if !ok {
			http.Error(w, `Не найден shortURL `+link, http.StatusBadRequest)
		} else {
			w.Header().Set("Location", url.OriginalURL)
			w.WriteHeader(http.StatusTemporaryRedirect)
		}
	case <-r.Context().Done():
		logging.S().Infow("Получение короткого URL прервано на клиентской стороне")
		w.WriteHeader(http.StatusGone)
	}
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
	defer lockCounter.Unlock()
	nextNum := currURLNum

	ownerID := r.Context().Value(auth.CPownerID)

	su := dbf.StorageURL{OriginalURL: url,
		UUID:     nextNum,
		OWNERID:  ownerID.(int64),
		ShortURL: strconv.FormatInt(nextNum, 36)}
	dbf.Store.DBFSaveLink(&su)
	if su.UUID == nextNum {
		w.WriteHeader(http.StatusCreated)
		currURLNum++
	} else {
		w.WriteHeader(http.StatusConflict)
	}
	w.Write([]byte(makeURL(su.UUID)))
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
	if url.URL == "" {
		http.Error(w, `Не указан URL`, http.StatusBadRequest)
		return
	}

	var surl outURL
	lockCounter.Lock()
	defer lockCounter.Unlock()
	nextNum := currURLNum

	ownerID := r.Context().Value(auth.CPownerID)

	su := dbf.StorageURL{UUID: nextNum,
		OriginalURL: url.URL,
		OWNERID:     ownerID.(int64),
		ShortURL:    strconv.FormatInt(nextNum, 36)}
	dbf.Store.DBFSaveLink(&su)
	if su.UUID == nextNum {
		w.WriteHeader(http.StatusCreated)
		currURLNum++
	} else {
		w.WriteHeader(http.StatusConflict)
	}
	var buf bytes.Buffer
	surl.Result = makeURL(su.UUID)

	jsonEncoder := json.NewEncoder(&buf)
	jsonEncoder.Encode(surl)
	w.Write(buf.Bytes())
}

type cmnURL struct {
	CorrelationID *string `json:"correlation_id,omitempty"`
	OriginalURL   *string `json:"original_url,omitempty"`
	ShortURL      *string `json:"short_url,omitempty"`
}

//type outURL2 struct {
//	CorrelationID string `json:"correlation_id"`
//}

func SetJSONBatchLinkf(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var urls []cmnURL
	if err := json.NewDecoder(r.Body).Decode(&urls); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(urls) == 0 {
		http.Error(w, `Не указано никаких данных`, http.StatusBadRequest)
		return
	}

	var surls []cmnURL
	lockCounter.Lock()
	defer lockCounter.Unlock()

	ownerID := r.Context().Value(auth.CPownerID)

	for _, url := range urls {
		if *url.OriginalURL != "" {
			shorturl := makeURL(currURLNum)
			surl := cmnURL{CorrelationID: url.CorrelationID, ShortURL: &shorturl}
			surls = append(surls, surl)
			su := dbf.StorageURL{UUID: currURLNum,
				OriginalURL: *url.OriginalURL,
				OWNERID:     ownerID.(int64),
				ShortURL:    strconv.FormatInt(currURLNum, 36)}
			dbf.Store.DBFSaveLink(&su)
			if su.UUID == currURLNum {
				currURLNum++
			}
		}
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(surls); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(buf.Bytes())
}

func GetOwnerURLsListf(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ownerValid := r.Context().Value(auth.CPownerValid)
	if !ownerValid.(bool) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var (
		ownerID int64
		urls    []dbf.StorageURL
		err     error
	)
	completed := make(chan struct{})

	go func() {
		ownerID = r.Context().Value(auth.CPownerID).(int64)
		urls, err = dbf.Store.DBFGetOwnURLs(ownerID)
		completed <- struct{}{}
	}()

	select {
	case <-completed:
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		links := make([]cmnURL, 0)
		for _, url := range urls {
			if url.OriginalURL != "" {
				link := cmnURL{}
				link.ShortURL = new(string)
				link.OriginalURL = new(string)
				*link.ShortURL = config.Addresses.Out.Host + ":" + strconv.Itoa(config.Addresses.Out.Port) + "/" + url.ShortURL
				*link.OriginalURL = url.OriginalURL
				links = append(links, link)
			}
		}

		if len(links) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(links); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Write(buf.Bytes())

	case <-r.Context().Done():
		logging.S().Infow("Получение списка URL для ownerID прервано на клиентской стороне")
		w.WriteHeader(http.StatusGone)
	}
}

func makeURL(num int64) string {
	if config.Addresses.In == nil {
		config.ReadData()
	}
	return config.Addresses.Out.Host + ":" + strconv.Itoa(config.Addresses.Out.Port) + "/" + strconv.FormatInt(num, 36)
}
