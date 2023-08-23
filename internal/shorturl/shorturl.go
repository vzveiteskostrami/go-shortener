package shorturl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
			if url.Deleted {
				w.WriteHeader(http.StatusGone)
			} else {
				w.Header().Set("Location", url.OriginalURL)
				w.WriteHeader(http.StatusTemporaryRedirect)
			}
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
	fmt.Fprintln(os.Stdout, "Записан URL", url)
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
	Deleted       *bool   `json:"deleted,omitempty"`
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

func DeleteOwnerURLsListf(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	var surls []string
	if err := json.NewDecoder(r.Body).Decode(&surls); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)

	ownerID := r.Context().Value(auth.CPownerID).(int64)

	go func() {
		doneCh := make(chan struct{})
		defer close(doneCh)

		dataCh := delRun(doneCh, surls)

		channels := delFanOut(doneCh, dataCh, len(surls), ownerID)

		outCh := delFanIn(doneCh, channels...)

		dbf.Store.BeginDel()
		for res := range outCh {
			if res != "" {
				dbf.Store.AddToDel(res)
			}
		}
		dbf.Store.EndDel()
	}()
}

func makeURL(num int64) string {
	if config.Addresses.In == nil {
		config.ReadData()
	}
	return config.Addresses.Out.Host + ":" + strconv.Itoa(config.Addresses.Out.Port) + "/" + strconv.FormatInt(num, 36)
}

func delRun(doneCh chan struct{}, input []string) chan string {
	inputCh := make(chan string)

	go func() {
		defer close(inputCh)

		for _, data := range input {
			select {
			case <-doneCh:
				return
			case inputCh <- data:
			}
		}
	}()

	return inputCh
}

func delFanOut(doneCh chan struct{}, inputCh chan string, sz int, ownerID int64) []chan string {
	// количество горутин add
	numWorkers := 10
	if sz < 10 {
		numWorkers = sz
	}
	// каналы, в которые отправляются результаты
	channels := make([]chan string, numWorkers)

	for i := 0; i < numWorkers; i++ {
		// получаем канал из горутины delCheck
		addResultCh := delCheck(doneCh, inputCh, ownerID)
		// отправляем его в слайс каналов
		channels[i] = addResultCh
	}

	// возвращаем слайс каналов
	return channels
}

func delCheck(doneCh chan struct{}, inputCh chan string, ownerID int64) chan string {
	addRes := make(chan string)

	go func() {
		defer close(addRes)

		for data := range inputCh {
			result := data
			if url, ok := dbf.Store.FindLink(data, true); ok {
				if url.Deleted || url.OWNERID != ownerID {
					result = ""
				}
			} else {
				result = ""
			}

			select {
			case <-doneCh:
				return
			case addRes <- result:
			}
		}
	}()
	return addRes
}

func delFanIn(doneCh chan struct{}, resultChs ...chan string) chan string {
	// конечный выходной канал в который отправляем данные из всех каналов из слайса, назовём его результирующим
	finalCh := make(chan string)

	// понадобится для ожидания всех горутин
	var wg sync.WaitGroup

	// перебираем все входящие каналы
	for _, ch := range resultChs {
		// в горутину передавать переменную цикла нельзя, поэтому делаем так
		chClosure := ch

		// инкрементируем счётчик горутин, которые нужно подождать
		wg.Add(1)

		go func() {
			// откладываем сообщение о том, что горутина завершилась
			defer wg.Done()

			// получаем данные из канала
			for data := range chClosure {
				select {
				// выходим из горутины, если канал закрылся
				case <-doneCh:
					return
				// если не закрылся, отправляем данные в конечный выходной канал
				case finalCh <- data:
				}
			}
		}()
	}

	go func() {
		// ждём завершения всех горутин
		wg.Wait()
		// когда все горутины завершились, закрываем результирующий канал
		close(finalCh)
	}()

	// возвращаем результирующий канал
	return finalCh
}
