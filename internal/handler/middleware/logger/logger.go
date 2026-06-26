package logger

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/paulwwyvern/gophermart/pkg/httphelpers/httperr"
	"github.com/paulwwyvern/gophermart/pkg/httphelpers/httpuser"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

func (w *responseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.size += size
	return size, err
}

// WithLogger логирует все запросы
func WithLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{ResponseWriter: w}

			next.ServeHTTP(rw, r)

			url := r.RequestURI
			method := r.Method
			userID := httpuser.GetUserID(r)
			err := httperr.GetError(r)

			logger.Info("http: Get Request",
				slog.Int64("user id", userID),
				slog.String("url", url),
				slog.String("method", method),
				slog.Int("status", rw.statusCode),
				slog.Int("size", rw.size),
				slog.Duration("duration", time.Since(start)),
				slog.Any("error", err),
			)
		})
	}
}
