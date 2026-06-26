package service

import (
	"context"

	"github.com/paulwwyvern/gophermart/internal/model"
	"github.com/paulwwyvern/gophermart/internal/model/dto"
	"github.com/paulwwyvern/gophermart/internal/model/errs"
	"github.com/paulwwyvern/gophermart/pkg/luhn"
	"github.com/shopspring/decimal"
)

type BalanceRepository interface {
	BeginTx(ctx context.Context) (context.Context, error)
	CommitTx(ctx context.Context) error
	RollbackTx(ctx context.Context) error

	AddUserBalanceByID(ctx context.Context, userId int64, add decimal.Decimal) error
	AddUserWithdrawnByID(ctx context.Context, userID int64, add decimal.Decimal) error
	GetUserBalanceByID(ctx context.Context, userId int64) (*model.UserBalance, error)
	CreateWithdrawal(ctx context.Context, withdrawal *model.Withdrawal) error
	GetWithdrawalsByUserID(ctx context.Context, userID int64) ([]*model.Withdrawal, error)
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

func (s *BalanceService) CreateWithdrawal(ctx context.Context, req *dto.CreateWithdrawalRequest) error {
	isNotValid := luhn.Validate(req.OrderNumber)
	if isNotValid != 0 {
		return errs.ErrOrderNumberUnprocessable
	}

	ctxTx, err := s.balanceRepository.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer s.balanceRepository.RollbackTx(ctxTx)

	// ПОлучаем текущий баланс
	userBalance, err := s.balanceRepository.GetUserBalanceByID(ctxTx, req.UserID)
	if err != nil {
		return err
	}

	// Если баллов недостаточно, то кидаем ошибку
	if userBalance.Balance.LessThan(req.Sum) {
		return errs.ErrBalanceNotEnough
	}

	// Снимаем деньги
	err = s.balanceRepository.AddUserBalanceByID(ctxTx, req.UserID, req.Sum.Neg())
	if err != nil {
		return err
	}

	// добавляем траты
	err = s.balanceRepository.AddUserWithdrawnByID(ctxTx, req.UserID, req.Sum)
	if err != nil {
		return err
	}

	withdrawal := &model.Withdrawal{
		UserID:      req.UserID,
		OrderNumber: req.OrderNumber,
		Sum:         req.Sum,
	}

	// сохраняем списание
	err = s.balanceRepository.CreateWithdrawal(ctxTx, withdrawal)
	if err != nil {
		return err
	}

	return s.balanceRepository.CommitTx(ctxTx)
}

func (s *BalanceService) GetWithdrawalsByUserID(ctx context.Context, userID int64) ([]*dto.GetWithdrawalsResponse, error) {
	withdrawals, err := s.balanceRepository.GetWithdrawalsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	res := make([]*dto.GetWithdrawalsResponse, 0, len(withdrawals))
	for _, withdrawal := range withdrawals {
		res = append(res, &dto.GetWithdrawalsResponse{
			OrderNumber: withdrawal.OrderNumber,
			Sum:         withdrawal.Sum,
			ProcessedAt: withdrawal.ProceedAt,
		})
	}

	return res, nil
}
