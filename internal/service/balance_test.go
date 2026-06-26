package service

import (
	"context"
	"errors"
	"testing"

	"github.com/paulwwyvern/gophermart/internal/model"
	"github.com/paulwwyvern/gophermart/internal/model/dto"
	"github.com/paulwwyvern/gophermart/internal/model/errs"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -source=balance.go -destination=mock_balance_test.go -package=service

func TestBalanceService_AddUserBalance(t *testing.T) {
	userId := int64(123456)
	add := decimal.NewFromFloat(3.14)

	ctrl := gomock.NewController(t)
	balanceRepository := NewMockBalanceRepository(ctrl)
	balanceRepository.EXPECT().AddUserBalanceByID(gomock.Any(), userId, add).Return(nil)

	service := NewBalanceService(balanceRepository)

	err := service.AddUserBalance(context.Background(), userId, add)
	assert.NoError(t, err)
}

func TestBalanceService_GetUserBalance(t *testing.T) {
	tests := []struct {
		name    string
		userId  int64
		balance *model.UserBalance
		err     error
		want    *dto.GetUserBalanceResponse
	}{
		{
			name:   "Test #1 success",
			userId: int64(123456),
			balance: &model.UserBalance{
				UserID:    int64(123456),
				Balance:   decimal.NewFromFloat(3.14),
				Withdrawn: decimal.NewFromFloat(2.71),
			},
			err: nil,
			want: &dto.GetUserBalanceResponse{
				Balance:   decimal.NewFromFloat(3.14),
				Withdrawn: decimal.NewFromFloat(2.71),
			},
		}, {
			name:    "Test #2 error",
			userId:  int64(123456),
			balance: &model.UserBalance{},
			err:     errors.New("internal error"),
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			balanceRepository := NewMockBalanceRepository(ctrl)
			balanceRepository.EXPECT().GetUserBalanceByID(gomock.Any(), tt.userId).Return(tt.balance, tt.err)
			service := NewBalanceService(balanceRepository)
			got, err := service.GetUserBalance(context.Background(), tt.userId)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.err, err)
		})
	}
}

func TestBalanceService_GetWithdrawalsByUserID(t *testing.T) {
	tests := []struct {
		name        string
		userId      int64
		withdrawals []*model.Withdrawal
		err         error
		want        []*dto.GetWithdrawalsResponse
	}{
		{
			name:   "Test #1 success",
			userId: int64(123456),
			withdrawals: []*model.Withdrawal{
				{
					WithdrawalID: int64(654321),
					UserID:       int64(123456),
					OrderNumber:  "abcdef",
					Sum:          decimal.NewFromFloat(3.14),
				},
				{
					WithdrawalID: int64(6543),
					UserID:       int64(1234),
					OrderNumber:  "123456",
					Sum:          decimal.NewFromFloat(2.71),
				},
			},
			err: nil,
			want: []*dto.GetWithdrawalsResponse{
				{
					OrderNumber: "abcdef",
					Sum:         decimal.NewFromFloat(3.14),
				},
				{
					OrderNumber: "123456",
					Sum:         decimal.NewFromFloat(2.71),
				},
			},
		}, {
			name:        "Test #2 error",
			userId:      int64(123456),
			withdrawals: []*model.Withdrawal{},
			err:         errors.New("internal error"),
			want:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			balanceRepository := NewMockBalanceRepository(ctrl)
			balanceRepository.EXPECT().GetWithdrawalsByUserID(gomock.Any(), tt.userId).Return(tt.withdrawals, tt.err)
			service := NewBalanceService(balanceRepository)
			got, err := service.GetWithdrawalsByUserID(context.Background(), tt.userId)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.err, err)
		})
	}
}

func TestBalanceService_CreateWithdrawalLuhn(t *testing.T) {
	req := dto.CreateWithdrawalRequest{
		UserID:      int64(123456),
		OrderNumber: "1234567890",
		Sum:         decimal.NewFromFloat(3.14),
	}
	ctrl := gomock.NewController(t)
	balanceRepository := NewMockBalanceRepository(ctrl)
	service := NewBalanceService(balanceRepository)
	err := service.CreateWithdrawal(context.Background(), &req)
	assert.ErrorIs(t, err, errs.ErrOrderNumberUnprocessable)
}

func TestBalanceService_CreateWithdrawalNotEnoughBalance(t *testing.T) {
	userId := int64(123456)

	req := dto.CreateWithdrawalRequest{
		UserID:      userId,
		OrderNumber: "12345678903",
		Sum:         decimal.NewFromFloat(3.14),
	}
	balance := &model.UserBalance{
		UserID:    userId,
		Balance:   decimal.NewFromFloat(2.17),
		Withdrawn: decimal.NewFromFloat(3.14),
	}

	ctx := context.Background()
	ctxTx := context.WithValue(ctx, "tx", "tx")

	ctrl := gomock.NewController(t)
	balanceRepository := NewMockBalanceRepository(ctrl)
	balanceRepository.EXPECT().BeginTx(ctx).Return(ctxTx, nil)
	balanceRepository.EXPECT().GetUserBalanceByID(ctxTx, userId).Return(balance, nil)
	balanceRepository.EXPECT().RollbackTx(ctxTx).Return(nil)

	service := NewBalanceService(balanceRepository)
	err := service.CreateWithdrawal(ctx, &req)
	assert.ErrorIs(t, err, errs.ErrBalanceNotEnough)
}

func TestBalanceService_CreateWithdrawalSuccess(t *testing.T) {
	userId := int64(123456)

	req := dto.CreateWithdrawalRequest{
		UserID:      userId,
		OrderNumber: "12345678903",
		Sum:         decimal.NewFromFloat(2.17),
	}
	balance := &model.UserBalance{
		UserID:    userId,
		Balance:   decimal.NewFromFloat(2.17),
		Withdrawn: decimal.NewFromFloat(3.14),
	}
	withdrawal := &model.Withdrawal{
		UserID:      userId,
		OrderNumber: "12345678903",
		Sum:         decimal.NewFromFloat(2.17),
	}

	ctx := context.Background()
	ctxTx := context.WithValue(ctx, "tx", "tx")

	ctrl := gomock.NewController(t)
	balanceRepository := NewMockBalanceRepository(ctrl)
	balanceRepository.EXPECT().BeginTx(ctx).Return(ctxTx, nil)
	balanceRepository.EXPECT().GetUserBalanceByID(ctxTx, userId).Return(balance, nil)
	balanceRepository.EXPECT().AddUserBalanceByID(ctxTx, userId, decimal.NewFromFloat(-2.17)).Return(nil)
	balanceRepository.EXPECT().AddUserWithdrawnByID(ctxTx, userId, decimal.NewFromFloat(2.17)).Return(nil)
	balanceRepository.EXPECT().CreateWithdrawal(ctxTx, withdrawal).Return(nil)
	balanceRepository.EXPECT().CommitTx(ctxTx).Return(nil)

	balanceRepository.EXPECT().RollbackTx(ctxTx).Return(nil)

	service := NewBalanceService(balanceRepository)
	err := service.CreateWithdrawal(ctx, &req)
	assert.NoError(t, err)
}
