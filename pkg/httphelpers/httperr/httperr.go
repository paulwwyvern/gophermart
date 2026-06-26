package httperr

import (
	"context"
	"net/http"
)

type ctxKey string

var key = ctxKey("httperr")

type HTTPHandler func(http.ResponseWriter, *http.Request) error

func Adapt(h HTTPHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err != nil {
			*r = *r.WithContext(context.WithValue(r.Context(), key, err))
		}
	}
}

func GetError(r *http.Request) error {
	err, ok := r.Context().Value(key).(error)
	if !ok {
		return nil
	}
	return err
}
