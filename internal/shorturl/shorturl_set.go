package shorturl

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/vzveiteskostrami/go-shortener/internal/auth"
	"github.com/vzveiteskostrami/go-shortener/internal/dbf"
)

func SetLink() http.Handler {
	return http.HandlerFunc(SetLinkf)
}

func SetLinkf(w http.ResponseWriter, r *http.Request) {
	// сохранён/закомментирован вывод на экран. Необходим для сложных случаев тестирования.
	//fmt.Fprintln(os.Stdout, "**************", "SetLinkf")
	w.Header().Set("Content-Type", "text/plain")
	b, err := io.ReadAll(r.Body)
	if err != nil {
		// сохранён/закомментирован вывод на экран. Необходим для сложных случаев тестирования.
		//fmt.Fprintln(os.Stdout, "Ошибка", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	url := string(b)
	if url == "" {
		// сохранён/закомментирован вывод на экран. Необходим для сложных случаев тестирования.
		//fmt.Fprintln(os.Stdout, "Ошибка не указан URL")
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
	lockWrite.Lock()
	defer lockWrite.Unlock()
	err = dbf.Store.DBFSaveLink(&su)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		if su.UUID == nextNum {
			w.WriteHeader(http.StatusCreated)
			currURLNum++
		} else {
			w.WriteHeader(http.StatusConflict)
		}
		w.Write([]byte(makeURL(su.UUID)))
	}
	// сохранён/закомментирован вывод на экран. Необходим для сложных случаев тестирования.
	//fmt.Fprintln(os.Stdout, "Записан URL", url, "как", su.ShortURL)
}

func SetJSONLink() http.Handler {
	return http.HandlerFunc(SetJSONLinkf)
}

func SetJSONLinkf(w http.ResponseWriter, r *http.Request) {
	// сохранён/закомментирован вывод на экран. Необходим для сложных случаев тестирования.
	//fmt.Fprintln(os.Stdout, "^^^^^^^^^^^^^^", "SetJSONLinkf")
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
	lockWrite.Lock()
	defer lockWrite.Unlock()
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

func SetJSONBatchLinkf(w http.ResponseWriter, r *http.Request) {
	// сохранён/закомментирован вывод на экран. Необходим для сложных случаев тестирования.
	//fmt.Fprintln(os.Stdout, "^^^^^^^^^^^^^^", "SetJSONBatchLinkf")
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
	lockWrite.Lock()
	defer lockWrite.Unlock()

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
