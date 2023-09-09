package logging

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type ContextParam struct {
	Completed bool
}

type ContextParamName string

const ContextCompletedKey ContextParamName = "completed"

var (
	sugar  zap.SugaredLogger
	logger *zap.Logger
)

func S() *zap.SugaredLogger {
	return &sugar
}

func LoggingInit() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	//defer logger.Sync()

	// делаем регистратор SugaredLogger
	sugar = *logger.Sugar()
}

func LoggingSync() {
	logger.Sync()
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

func WithLogging(h http.Handler) http.Handler {
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

		aa := ContextParam{Completed: true}
		ctx := context.WithValue(r.Context(), ContextCompletedKey, aa)
		// точка, где выполняется внутренний хендлер
		h.ServeHTTP(&lw, r.WithContext(ctx)) // обслуживание оригинального запроса

		// Since возвращает разницу во времени между start
		// и моментом вызова Since. Таким образом можно посчитать
		// время выполнения запроса.
		duration := time.Since(start)

		sugar.Infoln(
			"uri:", r.RequestURI,
			"method:", r.Method,
			"status:", responseData.status, http.StatusText(responseData.status),
			"duration:", duration,
			"size:", responseData.size,
		)
	}
	// возвращаем функционально расширенный хендлер
	return http.HandlerFunc(logFn)
}
