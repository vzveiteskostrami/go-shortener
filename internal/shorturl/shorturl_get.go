package shorturl

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/vzveiteskostrami/go-shortener/internal/auth"
	"github.com/vzveiteskostrami/go-shortener/internal/config"
	"github.com/vzveiteskostrami/go-shortener/internal/dbf"
)

func GetLink() http.Handler {
	return http.HandlerFunc(GetLinkf)
}

func GetLinkf(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	link := chi.URLParam(r, "shlink")

	url, err := dbf.Store.FindLink(r.Context(), link, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		if url.Deleted {
			w.WriteHeader(http.StatusGone)
		} else {
			w.Header().Set("Location", url.OriginalURL)
			w.WriteHeader(http.StatusTemporaryRedirect)
		}
	}
}

/*
func GetLinkf(w http.ResponseWriter, r *http.Request) {
	// сохранён/закомментирован вывод на экран. Необходим для сложных случаев тестирования.
	//fmt.Fprintln(os.Stdout, "##############", "GetLinkf")
	w.Header().Set("Content-Type", "text/plain")
	link := chi.URLParam(r, "shlink")

	completed := make(chan struct{})
	url := dbf.StorageURL{}
	ok := false

	go func() {
		url, ok = dbf.Store.FindLink(r.Context(), link, true)
		completed <- struct{}{}
	}()

	select {
	case <-completed:
		if !ok {
			http.Error(w, `Не найден shortURL `+link, http.StatusBadRequest)
		} else {
			//logging.S().Info("find LINK#", link)

			if url.Deleted {
				//logging.S().Info("find LINK# DELETED")
				w.WriteHeader(http.StatusGone)
			} else {
				//logging.S().Info("find LINK# GODNYY")
				w.Header().Set("Location", url.OriginalURL)
				w.WriteHeader(http.StatusTemporaryRedirect)
			}
		}
	case <-r.Context().Done():
		logging.S().Infow("Получение короткого URL прервано на клиентской стороне")
		w.WriteHeader(http.StatusGone)
	}
}
*/

func GetOwnerURLsListf(w http.ResponseWriter, r *http.Request) {
	// сохранён/закомментирован вывод на экран. Необходим для сложных случаев тестирования.
	///fmt.Fprintln(os.Stdout, "^^^^^^^^^^^^^^", "GetOwnerURLsListf")
	w.Header().Set("Content-Type", "application/json")

	var (
		ownerID int64
		urls    []dbf.StorageURL
		err     error
	)
	ownerID = r.Context().Value(auth.CPownerID).(int64)
	urls, err = dbf.Store.DBFGetOwnURLs(r.Context(), ownerID)

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
}

/*
func GetOwnerURLsListf(w http.ResponseWriter, r *http.Request) {
	// сохранён/закомментирован вывод на экран. Необходим для сложных случаев тестирования.
	///fmt.Fprintln(os.Stdout, "^^^^^^^^^^^^^^", "GetOwnerURLsListf")
	w.Header().Set("Content-Type", "application/json")

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
*/
