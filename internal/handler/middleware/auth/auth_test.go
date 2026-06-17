package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/paulwwyvern/gophermart/pkg/httphelpers/httpuser"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -source=auth.go -destination=mock_auth_test.go -package=auth

func echoUser(w http.ResponseWriter, r *http.Request) {

	user := httpuser.GetUserID(r)

	w.WriteHeader(200)
	w.Write([]byte(strconv.Itoa(int(user))))
}

func TestAuth_Success(t *testing.T) {
	token := "token"
	userId := int64(123456)

	ctrl := gomock.NewController(t)
	tokenValidator := NewMockTokenValidator(ctrl)
	tokenValidator.EXPECT().ValidateToken(token).Return(userId, nil)

	handler := WithAuth(tokenValidator)(http.HandlerFunc(echoUser))

	r := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	r.AddCookie(&http.Cookie{
		Name:  "Token",
		Value: token,
	})

	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, strconv.Itoa(int(userId)), w.Body.String())

}

func TestAuth_WithInvalidToken(t *testing.T) {
	token := "token"

	ctrl := gomock.NewController(t)
	tokenValidator := NewMockTokenValidator(ctrl)
	tokenValidator.EXPECT().ValidateToken(token).Return(int64(0), errors.New("invalid token"))

	handler := WithAuth(tokenValidator)(http.HandlerFunc(echoUser))

	r := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	r.AddCookie(&http.Cookie{
		Name:  "Token",
		Value: token,
	})

	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_WithoutToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	tokenValidator := NewMockTokenValidator(ctrl)

	handler := WithAuth(tokenValidator)(http.HandlerFunc(echoUser))

	r := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
