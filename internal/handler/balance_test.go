package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/paulwwyvern/gophermart/internal/model/dto"
	"github.com/paulwwyvern/gophermart/internal/model/errs"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -source=balance.go -destination=mock_balance_test.go -package=handler

func TestHandler_AddBalance(t *testing.T) {
	add := decimal.NewFromFloat(3.14)

	ctrl := gomock.NewController(t)
	userService := NewMockUserService(ctrl)
	balanceService := NewMockBalanceService(ctrl)
	orderService := NewMockOrderService(ctrl)

	balanceService.EXPECT().AddUserBalance(gomock.Any(), gomock.Any(), add).Return(nil)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("3.14"))
	w := httptest.NewRecorder()

	h := NewHandler(1024*1024, userService, orderService, balanceService)

	h.AddBalance(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_GetBalance(t *testing.T) {
	decimal.MarshalJSONWithoutQuotes = true

	balance := &dto.GetUserBalanceResponse{
		Balance:   decimal.NewFromFloat(3.14),
		Withdrawn: decimal.NewFromFloat(2.71),
	}

	ctrl := gomock.NewController(t)
	userService := NewMockUserService(ctrl)
	balanceService := NewMockBalanceService(ctrl)
	orderService := NewMockOrderService(ctrl)

	balanceService.EXPECT().GetUserBalance(gomock.Any(), gomock.Any()).Return(balance, nil)

	r := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	h := NewHandler(1024*1024, userService, orderService, balanceService)

	h.GetBalance(w, r)

	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"current":3.14,"withdrawn":2.71}`, string(w.Body.Bytes()))
}

func TestHandler_GetBalance_Internal(t *testing.T) {
	decimal.MarshalJSONWithoutQuotes = true

	balance := &dto.GetUserBalanceResponse{}

	ctrl := gomock.NewController(t)
	userService := NewMockUserService(ctrl)
	balanceService := NewMockBalanceService(ctrl)
	orderService := NewMockOrderService(ctrl)

	balanceService.EXPECT().GetUserBalance(gomock.Any(), gomock.Any()).Return(balance, errors.New("internal error"))

	r := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	h := NewHandler(1024*1024, userService, orderService, balanceService)

	h.GetBalance(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandler_CreateWithdrawal(t *testing.T) {
	decimal.MarshalJSONWithoutQuotes = true

	tests := []struct {
		name     string
		err      error
		wantCode int
	}{
		{
			name:     "Test #1: success",
			err:      nil,
			wantCode: http.StatusOK,
		},
		{
			name:     "Test #2: internal error",
			err:      errors.New("internal error"),
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "Test #3: Not Enough",
			err:      errs.ErrBalanceNotEnough,
			wantCode: http.StatusPaymentRequired,
		},
		{
			name:     "Test #2: wrong number",
			err:      errs.ErrOrderNumberUnprocessable,
			wantCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			withdraw := &dto.CreateWithdrawalRequest{
				UserID:      -1,
				OrderNumber: "12345",
				Sum:         decimal.NewFromFloat(2.71),
			}
			ctrl := gomock.NewController(t)
			userService := NewMockUserService(ctrl)
			balanceService := NewMockBalanceService(ctrl)
			orderService := NewMockOrderService(ctrl)

			balanceService.EXPECT().CreateWithdrawal(gomock.Any(), withdraw).Return(tt.err)

			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"order":"12345","sum":2.71}`))
			w := httptest.NewRecorder()

			h := NewHandler(1024*1024, userService, orderService, balanceService)

			h.CreateWithdrawal(w, r)

			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}

func TestHandler_GetWithdrawal(t *testing.T) {
	decimal.MarshalJSONWithoutQuotes = true

	withdrawals := []*dto.GetWithdrawalsResponse{
		{
			OrderNumber: "12345",
			Sum:         decimal.NewFromFloat(2.71),
		}, {
			OrderNumber: "67890",
			Sum:         decimal.NewFromFloat(3.14),
		},
	}

	ctrl := gomock.NewController(t)
	userService := NewMockUserService(ctrl)
	balanceService := NewMockBalanceService(ctrl)
	orderService := NewMockOrderService(ctrl)

	balanceService.EXPECT().GetWithdrawalsByUserID(gomock.Any(), gomock.Any()).Return(withdrawals, nil)

	r := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	h := NewHandler(1024*1024, userService, orderService, balanceService)

	h.GetWithdrawals(w, r)

	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `[{"order":"12345","sum":2.71,"processed_at":"0001-01-01T00:00:00Z"},{"order":"67890","sum":3.14,"processed_at":"0001-01-01T00:00:00Z"}]`, string(w.Body.Bytes()))
}

func TestHandler_GetWithdrawal_NoContent(t *testing.T) {
	decimal.MarshalJSONWithoutQuotes = true

	withdrawals := []*dto.GetWithdrawalsResponse{}

	ctrl := gomock.NewController(t)
	userService := NewMockUserService(ctrl)
	balanceService := NewMockBalanceService(ctrl)
	orderService := NewMockOrderService(ctrl)

	balanceService.EXPECT().GetWithdrawalsByUserID(gomock.Any(), gomock.Any()).Return(withdrawals, nil)

	r := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	h := NewHandler(1024*1024, userService, orderService, balanceService)

	h.GetWithdrawals(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestHandler_GetWithdrawal_Internal(t *testing.T) {
	decimal.MarshalJSONWithoutQuotes = true

	withdrawals := []*dto.GetWithdrawalsResponse{}

	ctrl := gomock.NewController(t)
	userService := NewMockUserService(ctrl)
	balanceService := NewMockBalanceService(ctrl)
	orderService := NewMockOrderService(ctrl)

	balanceService.EXPECT().GetWithdrawalsByUserID(gomock.Any(), gomock.Any()).Return(withdrawals, errors.New("internal error"))

	r := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	h := NewHandler(1024*1024, userService, orderService, balanceService)

	h.GetWithdrawals(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
