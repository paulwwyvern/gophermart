package auth

import (
	"github.com/paulwwyvern/gophermart/pkg/httphelpers/httperr"
	"github.com/paulwwyvern/gophermart/pkg/httphelpers/httpuser"
	"net/http"
	"strings"
)

type TokenValidator interface {
	ValidateToken(token string) (int64, error)
}

func WithAuth(validator TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(httperr.Adapt(func(w http.ResponseWriter, r *http.Request) error {
			token := r.Header.Get("Authorization")
			token = strings.TrimPrefix(token, "Bearer ")
			userId, err := validator.ValidateToken(token)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return err
			}
			httpuser.SetUserID(r, userId)

			next.ServeHTTP(w, r)
			return nil
		}))
	}
}
