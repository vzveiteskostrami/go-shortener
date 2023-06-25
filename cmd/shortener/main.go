package main

import (
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	config "github.com/vzveiteskostrami/go-shortener/cmd/shortener/Cfg"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

var (
	currURLNum  int64
	store       map[string]string
	lockCounter sync.Mutex
	srv         *http.Server
	sugar       zap.SugaredLogger
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
	store = make(map[string]string)

	srv = &http.Server{
		Addr:        config.Addresses.In.Host + ":" + strconv.Itoa(config.Addresses.In.Port),
		Handler:     mainRouter(),
		IdleTimeout: time.Second * 1,
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

	r.Handle(`/`, withLogging(setLink()))
	r.Handle("/{shlink}", withLogging(getLink()))
	//r.Get("/{shlink}", getLink())
	//r.Post("/", setLink)
	return r
}

func getLink() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `Ожидался метод `+http.MethodPost, http.StatusBadRequest)
			return
		}
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
	return http.HandlerFunc(fn)
}

func setLink() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, `Ожидался метод `+http.MethodGet, http.StatusBadRequest)
			return
		}
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
