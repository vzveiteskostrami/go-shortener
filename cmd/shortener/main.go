package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	config "github.com/vzveiteskostrami/go-shortener/cmd/shortener/Cfg"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type storageURL struct {
	UUID        int64  `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

var (
	currURLNum  int64
	store       map[string]storageURL
	lockCounter sync.Mutex
	srv         *http.Server
	sugar       zap.SugaredLogger
	fStore      *os.File
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	// делаем регистратор SugaredLogger
	sugar = *logger.Sugar()

	config.ReadData()
	currURLNum = 0
	store = make(map[string]storageURL)

	fmt.Println(config.FileStorage.FileName)
	srv = &http.Server{
		Addr:        config.Addresses.In.Host + ":" + strconv.Itoa(config.Addresses.In.Port),
		Handler:     mainRouter(),
		IdleTimeout: time.Second * 1,
	}

	if config.FileStorage.FileName != `` {
		s := filepath.Dir(config.FileStorage.FileName)
		if s != `` {
			err = os.MkdirAll(s, fs.ModeDir)
			if err != nil {
				sugar.Panic(err)
			}
		}
		fStore, err = os.OpenFile(config.FileStorage.FileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			sugar.Panic(err)
		}
		defer fStore.Close()
		readStoredData()
	}

	sugar.Infow(
		"Starting server",
		"addr", config.Addresses.In.Host+":"+strconv.Itoa(config.Addresses.In.Port),
	)
	sugar.Fatal(srv.ListenAndServe())
}

func closeServer() {
	time.Sleep(100 * time.Millisecond)
	srv.Close()
}

func mainRouter() chi.Router {
	r := chi.NewRouter()

	r.Handle(`/`, gzipHandle(withLogging(setLink())))
	r.Handle("/{shlink}", gzipHandle(withLogging(getLink())))
	r.Handle("/api/shorten", gzipHandle(withLogging(setJSONLink())))
	//r.Get("/{shlink}", getLink())
	//r.Post("/", setLink)
	return r
}

func getLink() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		//if r.Method != http.MethodGet {
		//	http.Error(w, `Ожидался метод `+http.MethodGet, http.StatusBadRequest)
		//	return
		//}
		w.Header().Set("Content-Type", "text/plain")
		link := chi.URLParam(r, "shlink")

		url, ok := store[link]
		if !ok {
			http.Error(w, `Не найден shortURL `+link, http.StatusBadRequest)
			return
		}

		w.Header().Set("Location", url.OriginalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
	return http.HandlerFunc(fn)
}

func setLink() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		//if r.Method != http.MethodPost {
		//	http.Error(w, `Ожидался метод `+http.MethodPost, http.StatusBadRequest)
		//	return
		//}
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
			currNum := currURLNum
			currURLNum++
			lockCounter.Unlock()
			saveLink(storageURL{OriginalURL: url,
				UUID:     currNum,
				ShortURL: strconv.FormatInt(currNum, 36)})
			w.WriteHeader(http.StatusCreated)

			if config.Addresses.In == nil {
				config.ReadData()
			}

			w.Write([]byte(makeURL(currNum)))
		}
	}
	return http.HandlerFunc(fn)
}

type inURL struct {
	URL string `json:"url"`
}

type outURL struct {
	Result string `json:"result"`
}

func setJSONLink() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		//if r.Method != http.MethodPost {
		//	http.Error(w, `Ожидался метод `+http.MethodPost, http.StatusBadRequest)
		//	return
		//}
		w.Header().Set("Content-Type", "application/json")
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var url inURL
		if err := json.NewDecoder(r.Body).Decode(&url); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if IsNil(url.URL) || url.URL == `` {
			http.Error(w, `Не указан URL`, http.StatusBadRequest)
			return
		}

		var surl outURL
		lockCounter.Lock()
		currNum := currURLNum
		currURLNum++
		lockCounter.Unlock()
		saveLink(storageURL{UUID: currNum,
			OriginalURL: url.URL,
			ShortURL:    strconv.FormatInt(currNum, 36)})
		w.WriteHeader(http.StatusCreated)

		if config.Addresses.In == nil {
			config.ReadData()
		}

		var buf bytes.Buffer

		surl.Result = makeURL(currNum)

		jsonEncoder := json.NewEncoder(&buf)
		jsonEncoder.Encode(surl)
		w.Write(buf.Bytes())
	}
	return http.HandlerFunc(fn)
}

func withLogging(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}

		// точка, где выполняется внутренний хендлер
		h.ServeHTTP(&lw, r) // обслуживание оригинального запроса

		// Since возвращает разницу во времени между start
		// и моментом вызова Since. Таким образом можно посчитать
		// время выполнения запроса.
		duration := time.Since(start)

		sugar.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status,
			"duration", duration,
			"size", responseData.size,
		)
	}
	// возвращаем функционально расширенный хендлер
	return http.HandlerFunc(logFn)
}

type (
	// берём структуру для хранения сведений об ответе
	responseData struct {
		status int
		size   int
	}

	// добавляем реализацию http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

func IsNil(obj interface{}) bool {
	if obj == nil {
		return true
	}

	objValue := reflect.ValueOf(obj)
	// проверяем, что тип значения ссылочный, то есть в принципе может быть равен nil
	if objValue.Kind() != reflect.Ptr {
		return false
	}
	// проверяем, что значение равно nil
	//  важно, что IsNil() вызывает панику, если value не является ссылочным типом. Поэтому всегда проверяйте на Kind()
	if objValue.IsNil() {
		return true
	}

	return false
}

func gzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// переменная reader будет равна r.Body или *gzip.Reader
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gzp, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			r.Body = gzp
			defer gzp.Close()
		}
		// проверяем, что клиент поддерживает gzip-сжатие
		// это упрощённый пример. В реальном приложении следует проверять все
		// значения r.Header.Values("Accept-Encoding") и разбирать строку
		// на составные части, чтобы избежать неожиданных результатов
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			// если gzip не поддерживается, передаём управление
			// дальше без изменений
			next.ServeHTTP(w, r)
			return
		}

		// создаём gzip.Writer поверх текущего w
		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		// передаём обработчику страницы переменную типа gzipWriter для вывода данных
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	// w.Writer будет отвечать за gzip-сжатие, поэтому пишем в него
	return w.Writer.Write(b)
}

func readStoredData() error {

	if IsNil(fStore) {
		return nil
	}

	scanner := bufio.NewScanner(fStore)
	sho := storageURL{}
	var err error

	if store == nil {
		store = make(map[string]storageURL)
	}

	for scanner.Scan() {
		fmt.Println(scanner.Text())
		data := scanner.Bytes()
		err = json.Unmarshal(data, &sho)
		if err != nil {
			return err
		}
		store[sho.ShortURL] = sho
		if currURLNum <= sho.UUID {
			currURLNum = sho.UUID + 1
		}
	}
	return nil
}

func saveLink(sho storageURL) {
	if store == nil {
		store = make(map[string]storageURL)
	}
	store[sho.ShortURL] = sho

	if fStore == nil {
		return
	}

	data, _ := json.Marshal(&sho)
	// добавляем перенос строки
	data = append(data, '\n')
	_, _ = fStore.Write(data)
}

func makeURL(num int64) string {
	return config.Addresses.Out.Host + ":" + strconv.Itoa(config.Addresses.Out.Port) + "/" + strconv.FormatInt(num, 36)
}
