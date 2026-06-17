package auth

import (
	"net/http"

	"github.com/paulwwyvern/gophermart/pkg/httphelpers/httperr"
	"github.com/paulwwyvern/gophermart/pkg/httphelpers/httpuser"
)

// TokenValidator валилирует передаваемый ему токен, если он валиден,то должен вернуть айди пользователя
type TokenValidator interface {
	ValidateToken(token string) (int64, error)
}

// WithAuth аутентифицирует пользователя по его токену, который он передаёт через куки
//
// Если токена нет или он невалидный, то пользователя дальше не пускает и возвращается 409
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
