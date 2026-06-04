package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/paulwwyvern/gophermart/internal/model/dto"
	"github.com/paulwwyvern/gophermart/internal/model/errs"
	"github.com/paulwwyvern/gophermart/pkg/httphelpers/httperr"
	"github.com/paulwwyvern/gophermart/pkg/passwordhash"
)

type UserService interface {
	RegisterUser(ctx context.Context, login string, password string) (string, error)
	LoginUser(ctx context.Context, login string, password string) (string, error)
}

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	httperr.Adapt(h.registerUser)(w, r)
}

func (h *Handler) registerUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	r.Body = http.MaxBytesReader(w, r.Body, h.maxBytesLength)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return err
		} else if errors.Is(err, context.DeadlineExceeded) {
			w.WriteHeader(http.StatusRequestTimeout)
			return err
		}
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	var req dto.RegisterUserRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}
	token, err := h.userService.RegisterUser(ctx, req.Login, req.Password)

	if err != nil {
		var passwdHashErr *passwordhash.Error

		if errors.As(err, &passwdHashErr) {
			w.WriteHeader(http.StatusBadRequest)
			return err
		} else if errors.Is(err, errs.ErrUserAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			return err
		}
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	cookie := &http.Cookie{
		Name:     "Token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
	}

	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusOK)
	return nil
}

func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	httperr.Adapt(h.loginUser)(w, r)
}

func (h *Handler) loginUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	r.Body = http.MaxBytesReader(w, r.Body, h.maxBytesLength)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return err
		} else if errors.Is(err, context.DeadlineExceeded) {
			w.WriteHeader(http.StatusRequestTimeout)
			return err
		}
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	var req dto.LoginUserRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	token, err := h.userService.LoginUser(ctx, req.Login, req.Password)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) || errors.Is(err, errs.ErrAuthFailed) {
			w.WriteHeader(http.StatusUnauthorized)
		}
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	cookie := &http.Cookie{
		Name:     "Token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
	}

	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusOK)
	return nil
}
