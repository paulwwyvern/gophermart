package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/paulwwyvern/gophermart/internal/model/dto"
	"github.com/paulwwyvern/gophermart/internal/model/errs"
	"github.com/paulwwyvern/gophermart/pkg/httphelpers/httperr"
	"github.com/paulwwyvern/gophermart/pkg/httphelpers/httpuser"
)

type OrderService interface {
	CreateOrder(ctx context.Context, userId int64, number string) error
	GetOrdersByUserID(ctx context.Context, userId int64) ([]*dto.GetOrdersResponse, error)
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	httperr.Adapt(h.createOrder)(w, r)
}

func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) error {
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

	userID := httpuser.GetUserID(r)
	number := strings.TrimSpace(string(body))

	err = h.orderService.CreateOrder(ctx, userID, number)
	if err != nil {
		if errors.Is(err, errs.ErrOrderAlreadyUploaded) {
			w.WriteHeader(http.StatusOK)
		} else if errors.Is(err, errs.ErrOrderConflict) {
			w.WriteHeader(http.StatusConflict)
		} else if errors.Is(err, errs.ErrOrderNumberInvalid) {
			w.WriteHeader(http.StatusBadRequest)
		} else if errors.Is(err, errs.ErrOrderNumberUnprocessable) {
			w.WriteHeader(http.StatusUnprocessableEntity)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return err
	}

	w.WriteHeader(http.StatusCreated)
	return nil
}

func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	httperr.Adapt(h.getOrders)(w, r)
}

func (h *Handler) getOrders(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	userID := httpuser.GetUserID(r)

	orders, err := h.orderService.GetOrdersByUserID(ctx, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(orders)
}
