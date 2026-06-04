package handler

import (
	"io"
	"net/http"
)

type Handler struct {
	maxBytesLength int64

	userService    UserService
	orderService   OrderService
	balanceService BalanceService
}

func NewHandler(maxBodyLength int64, userService UserService, orderService OrderService, balanceService BalanceService) *Handler {
	return &Handler{
		maxBytesLength: maxBodyLength,
		userService:    userService,
		orderService:   orderService,
		balanceService: balanceService,
	}
}

func (h *Handler) Echo(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
