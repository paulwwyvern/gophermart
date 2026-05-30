package handler

import (
	"io"
	"net/http"
)

type Handler struct {
	maxBytesLength int64

	userService UserService
}

func NewHandler(maxBodyLength int64, userService UserService) *Handler {
	return &Handler{
		maxBytesLength: maxBodyLength,
		userService:    userService,
	}
}

func (h *Handler) Echo(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
