package httpuser

import (
	"context"
	"net/http"
)

type ctxKey string

var key = ctxKey("httpuser")

func SetUserID(r *http.Request, userID int64) {
	*r = *r.WithContext(context.WithValue(r.Context(), key, userID))
}

func GetUserID(r *http.Request) int64 {
	userID, ok := r.Context().Value(key).(int64)
	if !ok {
		return -1
	}
	return userID
}
