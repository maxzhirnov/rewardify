package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// GzipMiddleware сжимает ответ, если клиент поддерживает gzip
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, поддерживает ли клиент gzip
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Создаем gzip writer
		gz := gzip.NewWriter(w)
		defer gz.Close()

		// Заменяем ResponseWriter для сжатия содержимого
		w.Header().Set("Content-Encoding", "gzip")
		gzipResponseWriter := &gzipWriter{Writer: gz, ResponseWriter: w}
		next.ServeHTTP(gzipResponseWriter, r)
	})
}

type gzipWriter struct {
	io.Writer
	http.ResponseWriter
}

// Переопределяем Write метод
func (w *gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
