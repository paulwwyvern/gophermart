package handler

import (
	"io"
	"net/http"
)

type Handler struct {
	userService UserService
}

func NewHandler(userService UserService) *Handler {
	return &Handler{
		userService: userService,
	}
}

func (h *Handler) Echo(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
