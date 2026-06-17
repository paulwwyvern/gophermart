package postgres

import (
	"context"
	"fmt"

	"github.com/paulwwyvern/gophermart/internal/model"
	"github.com/shopspring/decimal"
)

func (s *Storage) AddUserBalanceByID(ctx context.Context, userID int64, add decimal.Decimal) error {
	query := "UPDATE users SET balance = balance + $1 WHERE id = $2;"

	stmt, err := s.Prepare(ctx, query)
	if err != nil {
		return fmt.Errorf("Postgres.AddUserBalance: prepare statement error: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, add, userID)
	if err != nil {
		return fmt.Errorf("Postgres.AddUserBalance: exec error: %w", err)
	}
	return nil
}

func (s *Storage) AddUserWithdrawnByID(ctx context.Context, userID int64, add decimal.Decimal) error {
	query := "UPDATE users SET withdrawn = withdrawn + $1 WHERE id = $2;"

	stmt, err := s.Prepare(ctx, query)
	if err != nil {
		return fmt.Errorf("Postgres.AddUserWithdrawn: prepare statement error: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, add, userID)
	if err != nil {
		return fmt.Errorf("Postgres.AddUserWithdrawn: exec error: %w", err)
	}
	return nil
}

func (s *Storage) GetUserBalanceByID(ctx context.Context, userID int64) (*model.UserBalance, error) {
	query := "SELECT balance, withdrawn FROM users WHERE id = $1 FOR UPDATE;"

	stmt, err := s.Prepare(ctx, query)
	if err != nil {
		return &model.UserBalance{}, fmt.Errorf("Postgres.GetUserBalanceByLogin: prepare statement error: %w", err)
	}
	defer stmt.Close()

	balance := &model.UserBalance{}
	err = stmt.QueryRowContext(ctx, userID).Scan(&balance.Balance, &balance.Withdrawn)
	if err != nil {
		return &model.UserBalance{}, fmt.Errorf("Postgres.GetUserBalanceByLogin: query error: %w", err)
	}

	return balance, nil
}

func (s *Storage) CreateWithdrawal(ctx context.Context, withdrawal *model.Withdrawal) error {
	query := "INSERT INTO withdrawals(user_id, order_number, sum) VALUES($1, $2, $3);"

	stmt, err := s.Prepare(ctx, query)
	if err != nil {
		return fmt.Errorf("Postgres.CreateWithdrawal: prepare statement error: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, withdrawal.UserID, withdrawal.OrderNumber, withdrawal.Sum)
	if err != nil {
		return fmt.Errorf("Postgres.CreateWithdrawal: exec error: %w", err)
	}
	return nil
}

func (s *Storage) GetWithdrawalsByUserID(ctx context.Context, userID int64) ([]*model.Withdrawal, error) {
	query := "SELECT order_number, sum, processed_at FROM withdrawals WHERE user_id = $1 ORDER BY processed_at DESC;"

	stmt, err := s.Prepare(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("Postgres.GetWithdrawalsByUserID: prepare statement error: %w", err)
	}
	defer stmt.Close()

	var withdrawals []*model.Withdrawal
	rows, err := stmt.QueryContext(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("Postgres.GetWithdrawalsByUserID: query error: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		withdrawal := &model.Withdrawal{}
		err = rows.Scan(&withdrawal.OrderNumber, &withdrawal.Sum, &withdrawal.ProceedAt)
		if err != nil {
			return nil, fmt.Errorf("Postgres.GetWithdrawalsByUserID: rows.Scan error: %w", err)
		}

		withdrawals = append(withdrawals, withdrawal)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("Postgres.GetWithdrawalsByUserID: rows.Err() error: %w", err)
	}

	return withdrawals, nil

}
