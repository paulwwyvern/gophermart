package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/paulwwyvern/gophermart/internal/model/dto"
	"github.com/paulwwyvern/gophermart/pkg/httphelpers/httperr"
	"github.com/paulwwyvern/gophermart/pkg/httphelpers/httpuser"
	"github.com/shopspring/decimal"
)

type BalanceService interface {
	AddUserBalance(ctx context.Context, userId int64, add decimal.Decimal) error
	GetUserBalance(ctx context.Context, userId int64) (*dto.GetUserBalanceResponse, error)
}

func (h *Handler) AddBalance(w http.ResponseWriter, r *http.Request) {
	httperr.Adapt(h.addBalance)(w, r)
}

func (h *Handler) addBalance(w http.ResponseWriter, r *http.Request) error {
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
	bodyString := string(body)
	bodyString = strings.TrimSpace(bodyString)

	userID := httpuser.GetUserID(r)
	balance, err := decimal.NewFromString(bodyString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	err = h.balanceService.AddUserBalance(ctx, userID, balance)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	httperr.Adapt(h.getBalance)(w, r)
}

func (h *Handler) getBalance(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	userID := httpuser.GetUserID(r)
	balance, err := h.balanceService.GetUserBalance(ctx, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	w.WriteHeader(http.StatusOK)

	return json.NewEncoder(w).Encode(balance)
}
