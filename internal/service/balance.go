package service

import (
	"context"

	"github.com/paulwwyvern/gophermart/internal/model"
	"github.com/paulwwyvern/gophermart/internal/model/dto"
	"github.com/shopspring/decimal"
)

type BalanceRepository interface {
	AddUserBalanceByID(ctx context.Context, userId int64, add decimal.Decimal) error
	GetUserBalanceByID(ctx context.Context, userId int64) (*model.UserBalance, error)
}

type BalanceService struct {
	balanceRepository BalanceRepository
}

func NewBalanceService(balanceRepository BalanceRepository) *BalanceService {
	return &BalanceService{
		balanceRepository: balanceRepository,
	}
}

func (s *BalanceService) AddUserBalance(ctx context.Context, userId int64, add decimal.Decimal) error {
	return s.balanceRepository.AddUserBalanceByID(ctx, userId, add)
}

func (s *BalanceService) GetUserBalance(ctx context.Context, userId int64) (*dto.GetUserBalanceResponse, error) {
	balance, err := s.balanceRepository.GetUserBalanceByID(ctx, userId)
	if err != nil {
		return nil, err
	}

	res := &dto.GetUserBalanceResponse{
		Balance:   balance.Balance,
		Withdrawn: balance.Withdrawn,
	}
	return res, nil
}
