package auth

import (
	"net/http"

	"github.com/paulwwyvern/gophermart/pkg/httphelpers/httperr"
	"github.com/paulwwyvern/gophermart/pkg/httphelpers/httpuser"
)

type TokenValidator interface {
	ValidateToken(token string) (int64, error)
}

func WithAuth(validator TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return httperr.Adapt(func(w http.ResponseWriter, r *http.Request) error {

			cookie, err := r.Cookie("Token")
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return err
			}

			token := cookie.Value

			userId, err := validator.ValidateToken(token)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return err
			}
			httpuser.SetUserID(r, userId)

			next.ServeHTTP(w, r)
			return nil
		})
	}
}
