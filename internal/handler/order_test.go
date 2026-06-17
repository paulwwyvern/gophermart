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

//go:generate mockgen -source=order.go -destination=mock_order_test.go -package=handler

func TestHandler_CreateOrder(t *testing.T) {
	decimal.MarshalJSONWithoutQuotes = true

	tests := []struct {
		name     string
		err      error
		wantCode int
	}{
		{
			name:     "Test #1: success",
			err:      nil,
			wantCode: http.StatusAccepted,
		},
		{
			name:     "Test #2: internal error",
			err:      errors.New("internal error"),
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "Test #3: Order already uploaded by this user",
			err:      errs.ErrOrderAlreadyUploaded,
			wantCode: http.StatusOK,
		},
		{
			name:     "Test #4: Order already uploaded by another user",
			err:      errs.ErrOrderConflict,
			wantCode: http.StatusConflict,
		},
		{
			name:     "Test #5: invalid number",
			err:      errs.ErrOrderNumberInvalid,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "Test #6: unprocessable number",
			err:      errs.ErrOrderNumberUnprocessable,
			wantCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			orderNumber := "12345678903"

			ctrl := gomock.NewController(t)
			userService := NewMockUserService(ctrl)
			balanceService := NewMockBalanceService(ctrl)
			orderService := NewMockOrderService(ctrl)

			orderService.EXPECT().CreateOrder(gomock.Any(), gomock.Any(), orderNumber).Return(tt.err)

			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`12345678903`))
			w := httptest.NewRecorder()

			h := NewHandler(1024*1024, userService, orderService, balanceService)

			h.CreateOrder(w, r)

			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}

func TestHandler_GetOrders(t *testing.T) {
	decimal.MarshalJSONWithoutQuotes = true

	orders := []*dto.GetOrdersResponse{
		{
			Number: "12345",
			Status: "NEW",
		}, {
			Number:  "67890",
			Status:  "PROCESSED",
			Accrual: decimal.NewFromFloat(3.14),
		},
	}

	ctrl := gomock.NewController(t)
	userService := NewMockUserService(ctrl)
	balanceService := NewMockBalanceService(ctrl)
	orderService := NewMockOrderService(ctrl)

	orderService.EXPECT().GetOrdersByUserID(gomock.Any(), gomock.Any()).Return(orders, nil)

	r := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	h := NewHandler(1024*1024, userService, orderService, balanceService)

	h.GetOrders(w, r)

	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `[{"number":"12345","status":"NEW","uploaded_at":"0001-01-01T00:00:00Z"},{"number":"67890","status":"PROCESSED","accrual":3.14,"uploaded_at":"0001-01-01T00:00:00Z"}]`, string(w.Body.Bytes()))
}

func TestHandler_GetOrders_NoContent(t *testing.T) {
	decimal.MarshalJSONWithoutQuotes = true

	orders := []*dto.GetOrdersResponse{}

	ctrl := gomock.NewController(t)
	userService := NewMockUserService(ctrl)
	balanceService := NewMockBalanceService(ctrl)
	orderService := NewMockOrderService(ctrl)

	orderService.EXPECT().GetOrdersByUserID(gomock.Any(), gomock.Any()).Return(orders, nil)

	r := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	h := NewHandler(1024*1024, userService, orderService, balanceService)

	h.GetOrders(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestHandler_GetOrders_Internal(t *testing.T) {
	decimal.MarshalJSONWithoutQuotes = true

	orders := []*dto.GetOrdersResponse{}

	ctrl := gomock.NewController(t)
	userService := NewMockUserService(ctrl)
	balanceService := NewMockBalanceService(ctrl)
	orderService := NewMockOrderService(ctrl)

	orderService.EXPECT().GetOrdersByUserID(gomock.Any(), gomock.Any()).Return(orders, errors.New("internal error"))

	r := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	h := NewHandler(1024*1024, userService, orderService, balanceService)

	h.GetOrders(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
