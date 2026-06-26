package service

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/paulwwyvern/gophermart/internal/model"
	"github.com/paulwwyvern/gophermart/internal/model/dto"
	"github.com/paulwwyvern/gophermart/internal/model/errs"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -source=order.go -destination=mock_order_test.go -package=service

func TestOrderService_CreateOrderLuhn(t *testing.T) {
	tests := []struct {
		name   string
		userId int64
		number string
		err    error
	}{
		{
			name:   "Test #1: invalid",
			userId: int64(123456),
			number: "asdgf",
			err:    errs.ErrOrderNumberInvalid,
		},
		{
			name:   "Test #2: unprocessable",
			userId: int64(123456),
			number: "1234567890",
			err:    errs.ErrOrderNumberUnprocessable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			orderRepo := NewMockOrderRepository(ctrl)

			service := NewOrderService(orderRepo)
			err := service.CreateOrder(context.Background(), tt.userId, tt.number)
			assert.Equal(t, tt.err, err)
		})
	}
}

func TestOrderService_CreateOrder(t *testing.T) {
	tests := []struct {
		name        string
		userId      int64
		number      string
		orderUserId int64
		err         error
		wantErr     error
	}{
		{
			name:        "Test #1: success",
			userId:      int64(123456),
			number:      "12345678903",
			orderUserId: int64(0),
			err:         nil,
			wantErr:     nil,
		},
		{
			name:        "Test #2: already uploaded",
			userId:      int64(123456),
			number:      "12345678903",
			orderUserId: int64(123456),
			err:         errs.ErrOrderAlreadyExists,
			wantErr:     errs.ErrOrderAlreadyUploaded,
		},
		{
			name:        "Test #3: conflict",
			userId:      int64(123456),
			number:      "12345678903",
			orderUserId: int64(12345),
			err:         errs.ErrOrderAlreadyExists,
			wantErr:     errs.ErrOrderConflict,
		},
		{
			name:        "Test #4: internal server error",
			userId:      int64(123456),
			number:      "12345678903",
			orderUserId: int64(12345),
			err:         io.EOF,
			wantErr:     io.EOF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			orderRepo := NewMockOrderRepository(ctrl)
			orderRepo.EXPECT().CreateOrder(gomock.Any(), tt.userId, tt.number).Return(tt.orderUserId, tt.err)

			service := NewOrderService(orderRepo)
			err := service.CreateOrder(context.Background(), tt.userId, tt.number)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestOrderService_GetOrdersByUserID(t *testing.T) {
	tests := []struct {
		name    string
		userId  int64
		orders  []*model.Order
		want    []*dto.GetOrdersResponse
		wantErr error
	}{
		{
			name:   "Test #1: success",
			userId: int64(123456),
			orders: []*model.Order{
				{
					Number:  "123456",
					Status:  "NEW",
					Accrual: decimal.NewFromFloat(3.14),
				},
				{
					Number:  "654321",
					Status:  "PROCESSED",
					Accrual: decimal.NewFromFloat(2.71),
				},
			},
			want: []*dto.GetOrdersResponse{
				{
					Number:  "123456",
					Status:  "NEW",
					Accrual: decimal.NewFromFloat(3.14),
				},
				{
					Number:  "654321",
					Status:  "PROCESSED",
					Accrual: decimal.NewFromFloat(2.71),
				},
			},
			wantErr: nil,
		},
		{
			name:    "Test #2: error",
			userId:  int64(123456),
			orders:  []*model.Order{},
			want:    nil,
			wantErr: errors.New("error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			orderRepo := NewMockOrderRepository(ctrl)
			orderRepo.EXPECT().GetOrdersByUserID(gomock.Any(), tt.userId).Return(tt.orders, tt.wantErr)
			service := NewOrderService(orderRepo)

			got, err := service.GetOrdersByUserID(context.Background(), tt.userId)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
