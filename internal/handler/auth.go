package handler

import (
	"context"
	"encoding/json"
	"github.com/paulwwyvern/gophermart/internal/model/dto"
	"github.com/paulwwyvern/gophermart/pkg/httphelpers/httperr"
	"io"
	"net/http"
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

	body, _ := io.ReadAll(r.Body)

	var req dto.RegisterUserRequest
	json.Unmarshal(body, &req)

	token, err := h.userService.RegisterUser(ctx, req.Login, req.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	w.Header().Set("Authorization", "Bearer "+token)

	w.WriteHeader(http.StatusOK)
	return nil
}

func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	httperr.Adapt(h.loginUser)(w, r)
}

func (h *Handler) loginUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	body, _ := io.ReadAll(r.Body)

	var req dto.LoginUserRequest
	json.Unmarshal(body, &req)

	token, err := h.userService.LoginUser(ctx, req.Login, req.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	w.Header().Set("Authorization", "Bearer "+token)

	w.WriteHeader(http.StatusOK)
	return nil
}
