package postgres

import (
	"context"
	"fmt"

	"github.com/paulwwyvern/gophermart/internal/model"
	"github.com/shopspring/decimal"
)

func (s *Storage) AddUserBalanceByID(ctx context.Context, userID int64, add decimal.Decimal) error {
	stmt, err := s.db.PrepareContext(ctx, "UPDATE users SET balance = balance + $1 WHERE id = $2;")
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

func (s *Storage) GetUserBalanceByID(ctx context.Context, userID int64) (*model.UserBalance, error) {
	stmt, err := s.db.PrepareContext(ctx, `SELECT balance, withdrawn FROM users WHERE id = $1;`)
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
