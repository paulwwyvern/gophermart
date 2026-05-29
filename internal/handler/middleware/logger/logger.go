package logger

import (
	"github.com/paulwwyvern/gophermart/pkg/httphelpers/httperr"
	"github.com/paulwwyvern/gophermart/pkg/httphelpers/httpuser"
	"go.uber.org/zap"
	"net/http"
	"time"
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

func WithLogger(logger *zap.Logger) func(http.Handler) http.Handler {
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
				zap.Int64("user id", userID),
				zap.String("url", url),
				zap.String("method", method),
				zap.Int("status", rw.statusCode),
				zap.Int("size", rw.size),
				zap.Duration("duration", time.Since(start)),
				zap.Error(err),
			)
		})
	}
}
