package compressing

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

func GZIPHandle(next http.Handler) http.Handler {
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

		gwr := gzipWriter{ResponseWriter: w, Writer: gz}
		defer func() {
			if gwr.Header().Get("Content-Encoding") == "gzip" {
				gz.Close()
			}
		}()

		// передаём обработчику страницы переменную типа gzipWriter для вывода данных
		next.ServeHTTP(gwr, r)
	})
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (gw gzipWriter) Write(b []byte) (int, error) {
	if gw.Header().Get("Content-Encoding") == "gzip" {
		// gw.Writer будет отвечать за gzip-сжатие, поэтому пишем в него
		return gw.Writer.Write(b)
	} else {
		return gw.ResponseWriter.Write(b)
	}
}

func (gw gzipWriter) WriteHeader(statusCode int) {
	ct := gw.Header().Get("Content-Type")
	if ct == "application/json" || ct == "text/html" {
		gw.Header().Set("Content-Encoding", "gzip")
	}
	gw.ResponseWriter.WriteHeader(statusCode)
}
