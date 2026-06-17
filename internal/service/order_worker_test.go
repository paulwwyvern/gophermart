package service

import (
	"context"
	"testing"

	"github.com/paulwwyvern/gophermart/internal/model"
	"github.com/paulwwyvern/gophermart/internal/model/errs"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestOrderWorkerPool_GetOrderStatus(t *testing.T) {
	ctx := context.Background()
	ctxTx := context.WithValue(ctx, "tx", "tx")
	logger := zap.NewNop()

	orders := []*model.ProcessingOrder{
		{
			UserID:      1,
			OrderNumber: "111",
		},
		{
			UserID:      2,
			OrderNumber: "222",
		},
		{
			UserID:      3,
			OrderNumber: "333",
		},
		{
			UserID:      4,
			OrderNumber: "444",
		},
		{
			UserID:      5,
			OrderNumber: "555",
		},
	}

	ctrl := gomock.NewController(t)
	accrualRepo := NewMockOrderAccrualRepository(ctrl)
	statusRepo := NewMockOrderStatusRepository(ctrl)

	gomock.InOrder(
		statusRepo.EXPECT().BeginTx(ctx).Return(ctxTx, nil),
		statusRepo.EXPECT().GetNewOrProcessingOrderNumbers(ctxTx, 4).Return(orders, nil),
		accrualRepo.EXPECT().GetOrderStatus(ctxTx, "111").Return("REGISTERED", decimal.Zero, nil),
		statusRepo.EXPECT().SetOrderStatus(ctxTx, "111", "PROCESSING", decimal.Zero).Return(nil),
		accrualRepo.EXPECT().GetOrderStatus(ctxTx, "222").Return("PROCESSING", decimal.Zero, nil),
		statusRepo.EXPECT().SetOrderStatus(ctxTx, "222", "PROCESSING", decimal.Zero).Return(nil),
		accrualRepo.EXPECT().GetOrderStatus(ctxTx, "333").Return("", decimal.Zero, errs.ErrAccrualNotRegistered),
		accrualRepo.EXPECT().GetOrderStatus(ctxTx, "444").Return("INVALID", decimal.Zero, nil),
		statusRepo.EXPECT().SetOrderStatus(ctxTx, "444", "INVALID", decimal.Zero).Return(nil),
		accrualRepo.EXPECT().GetOrderStatus(ctxTx, "555").Return("PROCESSED", decimal.NewFromFloat(3.14), nil),
		statusRepo.EXPECT().AddUserBalanceByID(ctxTx, int64(5), decimal.NewFromFloat(3.14)).Return(nil),
		statusRepo.EXPECT().SetOrderStatus(ctxTx, "555", "PROCESSED", decimal.NewFromFloat(3.14)).Return(nil),
		statusRepo.EXPECT().CommitTx(ctxTx).Return(nil),
		statusRepo.EXPECT().RollbackTx(ctxTx).Return(nil),
	)

	worker := NewOrderWorkerPool(accrualRepo, statusRepo, 1, 1, 1)

	err := worker.getOrderStatus(ctx, logger, 4)

	assert.NoError(t, err)
}

func TestOrderWorkerPool_GetOrderStatus_RateLimit(t *testing.T) {
	ctx := context.Background()
	ctxTx := context.WithValue(ctx, "tx", "tx")
	logger := zap.NewNop()

	orders := []*model.ProcessingOrder{
		{
			UserID:      1,
			OrderNumber: "111",
		},
		{
			UserID:      2,
			OrderNumber: "222",
		},
		{
			UserID:      3,
			OrderNumber: "333",
		},
	}

	ctrl := gomock.NewController(t)
	accrualRepo := NewMockOrderAccrualRepository(ctrl)
	statusRepo := NewMockOrderStatusRepository(ctrl)

	gomock.InOrder(
		statusRepo.EXPECT().BeginTx(ctx).Return(ctxTx, nil),
		statusRepo.EXPECT().GetNewOrProcessingOrderNumbers(ctxTx, 4).Return(orders, nil),
		accrualRepo.EXPECT().GetOrderStatus(ctxTx, "111").Return("REGISTERED", decimal.Zero, nil),
		statusRepo.EXPECT().SetOrderStatus(ctxTx, "111", "PROCESSING", decimal.Zero).Return(nil),
		accrualRepo.EXPECT().GetOrderStatus(ctxTx, "222").Return("", decimal.Zero,
			&errs.ErrAccrualTooManyRequests{RetryAfter: 1234, Body: "error"}),
		statusRepo.EXPECT().CommitTx(ctxTx).Return(nil),
		statusRepo.EXPECT().RollbackTx(ctxTx).Return(nil),
	)

	worker := NewOrderWorkerPool(accrualRepo, statusRepo, 1, 1, 1)

	err := worker.getOrderStatus(ctx, logger, 4)

	var errTooManyRequests *errs.ErrAccrualTooManyRequests

	assert.ErrorAs(t, err, &errTooManyRequests)

	assert.Equal(t, 1234, errTooManyRequests.RetryAfter)
	assert.Equal(t, "error", errTooManyRequests.Body)
}
