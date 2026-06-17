package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/paulwwyvern/gophermart/internal/model/dto"
	"github.com/paulwwyvern/gophermart/internal/model/errs"
	"github.com/paulwwyvern/gophermart/pkg/httphelpers/httperr"
	"github.com/paulwwyvern/gophermart/pkg/httphelpers/httpuser"
	"github.com/shopspring/decimal"
)

type BalanceService interface {
	AddUserBalance(ctx context.Context, userId int64, add decimal.Decimal) error
	GetUserBalance(ctx context.Context, userId int64) (*dto.GetUserBalanceResponse, error)
	CreateWithdrawal(ctx context.Context, req *dto.CreateWithdrawalRequest) error
	GetWithdrawalsByUserID(ctx context.Context, userID int64) ([]*dto.GetWithdrawalsResponse, error)
}

func (h *Handler) AddBalance(w http.ResponseWriter, r *http.Request) {
	httperr.Adapt(h.addBalance)(w, r)
}

func (h *Handler) addBalance(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	body, err := ReadBody(w, r, h.maxBytesLength)
	if err != nil {
		return err
	}

	body = bytes.TrimSpace(body)

	userID := httpuser.GetUserID(r)
	balance, err := decimal.NewFromString(string(body))
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	return json.NewEncoder(w).Encode(balance)
}

func (h *Handler) CreateWithdrawal(w http.ResponseWriter, r *http.Request) {
	httperr.Adapt(h.createWithdrawals)(w, r)
}

func (h *Handler) createWithdrawals(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	body, err := ReadBody(w, r, h.maxBytesLength)
	if err != nil {
		return err
	}

	userID := httpuser.GetUserID(r)

	req := &dto.CreateWithdrawalRequest{}
	err = json.Unmarshal(body, req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}
	req.UserID = userID

	err = h.balanceService.CreateWithdrawal(ctx, req)
	if err != nil {
		if errors.Is(err, errs.ErrBalanceNotEnough) {
			w.WriteHeader(http.StatusPaymentRequired)
			return err
		} else if errors.Is(err, errs.ErrOrderNumberUnprocessable) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return err
		}
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (h *Handler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	httperr.Adapt(h.getWithdrawals)(w, r)
}

func (h *Handler) getWithdrawals(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	userID := httpuser.GetUserID(r)

	withdrawals, err := h.balanceService.GetWithdrawalsByUserID(ctx, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(withdrawals)
}
