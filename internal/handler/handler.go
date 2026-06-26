package handler

import (
	"context"
	"errors"
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

func ReadBody(w http.ResponseWriter, r *http.Request, maxBytesLength int64) ([]byte, error) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBytesLength)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return nil, err
		} else if errors.Is(err, context.DeadlineExceeded) {
			w.WriteHeader(http.StatusRequestTimeout)
			return nil, err
		}
		w.WriteHeader(http.StatusBadRequest)
		return nil, err
	}
	return body, nil
}
