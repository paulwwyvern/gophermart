package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paulwwyvern/gophermart/internal/model/errs"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -source=auth.go -destination=mock_auth_test.go -package=handler

func TestHandler_RegisterUser(t *testing.T) {
	tests := []struct {
		name       string
		login      string
		password   string
		token      string
		err        error
		wantCookie bool
		wantCode   int
	}{
		{
			name:       "Test #1: Success",
			login:      "abcdef",
			password:   "123456",
			token:      "token",
			err:        nil,
			wantCode:   http.StatusOK,
			wantCookie: true,
		},
		{
			name:       "Test #2: Already Exists",
			login:      "abcdef",
			password:   "123456",
			token:      "",
			err:        errs.ErrUserAlreadyExists,
			wantCode:   http.StatusConflict,
			wantCookie: false,
		},
		{
			name:       "Test #3: Internal Server Error",
			login:      "abcdef",
			password:   "123456",
			token:      "",
			err:        errors.New("some error"),
			wantCode:   http.StatusInternalServerError,
			wantCookie: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			userService := NewMockUserService(ctrl)
			balanceService := NewMockBalanceService(ctrl)
			orderService := NewMockOrderService(ctrl)

			userService.EXPECT().RegisterUser(gomock.Any(), tt.login, tt.password).Return(tt.token, tt.err)

			bodyStruct := struct {
				Login    string `json:"login"`
				Password string `json:"password"`
			}{
				Login:    tt.login,
				Password: tt.password,
			}
			body, _ := json.Marshal(bodyStruct)
			bodyReader := bytes.NewReader(body)
			r := httptest.NewRequest(http.MethodGet, "/", bodyReader)
			w := httptest.NewRecorder()

			h := NewHandler(1024*1024, userService, orderService, balanceService)
			h.RegisterUser(w, r)

			assert.Equal(t, tt.wantCode, w.Code)
			if tt.wantCookie == true {
				cookies := w.Result().Cookies()
				for _, cookie := range cookies {
					if cookie.Name == "Token" {
						assert.Equal(t, tt.token, cookie.Value)
					}
				}
			}

		})
	}
}

func TestHandler_LoginUser(t *testing.T) {
	tests := []struct {
		name       string
		login      string
		password   string
		token      string
		err        error
		wantCookie bool
		wantCode   int
	}{
		{
			name:       "Test #1: Success",
			login:      "abcdef",
			password:   "123456",
			token:      "token",
			err:        nil,
			wantCode:   http.StatusOK,
			wantCookie: true,
		},
		{
			name:       "Test #2: Not Found",
			login:      "abcdef",
			password:   "123456",
			token:      "",
			err:        errs.ErrUserNotFound,
			wantCode:   http.StatusUnauthorized,
			wantCookie: false,
		},
		{
			name:       "Test #2: Authentication Failed",
			login:      "abcdef",
			password:   "123456",
			token:      "",
			err:        errs.ErrAuthFailed,
			wantCode:   http.StatusUnauthorized,
			wantCookie: false,
		},
		{
			name:       "Test #2: Internal Server Error",
			login:      "abcdef",
			password:   "123456",
			token:      "",
			err:        errors.New("some error"),
			wantCode:   http.StatusInternalServerError,
			wantCookie: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			userService := NewMockUserService(ctrl)
			balanceService := NewMockBalanceService(ctrl)
			orderService := NewMockOrderService(ctrl)

			userService.EXPECT().LoginUser(gomock.Any(), tt.login, tt.password).Return(tt.token, tt.err)

			bodyStruct := struct {
				Login    string `json:"login"`
				Password string `json:"password"`
			}{
				Login:    tt.login,
				Password: tt.password,
			}
			body, _ := json.Marshal(bodyStruct)
			bodyReader := bytes.NewReader(body)
			r := httptest.NewRequest(http.MethodGet, "/", bodyReader)
			w := httptest.NewRecorder()

			h := NewHandler(1024*1024, userService, orderService, balanceService)
			h.LoginUser(w, r)

			assert.Equal(t, tt.wantCode, w.Code)
			if tt.wantCookie == true {
				cookies := w.Result().Cookies()
				for _, cookie := range cookies {
					if cookie.Name == "Token" {
						assert.Equal(t, tt.token, cookie.Value)
					}
				}
			}

		})
	}
}
