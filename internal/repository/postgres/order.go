package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/paulwwyvern/gophermart/internal/model"
	"github.com/paulwwyvern/gophermart/internal/model/errs"
)

func (s *Storage) CreateOrder(ctx context.Context, userId int64, number string) (int64, error) {
	stmt, err := s.db.PrepareContext(ctx, "INSERT INTO orders (user_id, number) VALUES ($1, $2) RETURNING order_id")
	if err != nil {
		return -1, fmt.Errorf("Postgres.CreateOrder: prepare statement error: %w", err)
	}

	var orderId int64
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, userId, number).Scan(&orderId)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return -1, errs.ErrOrderAlreadyExists
			}
		}
		return -1, fmt.Errorf("Postgres.CreateOrder: exec query error: %w", err)
	}

	return orderId, nil
}

func (s *Storage) GetOrderByID(ctx context.Context, orderId int64) (*model.Order, error) {
	stmt, err := s.db.PrepareContext(ctx,
		`SELECT o.id, o.user_id, o.number, s.status, o.accrual, o.uploaded_at 
			FROM o orders JOIN s order_statuses ON o.status = s.id WHERE order_id = $1`)
	if err != nil {
		return nil, fmt.Errorf("Postgres.GetOrderByID: prepare statement error: %w", err)
	}
	defer stmt.Close()

	var order model.Order
	err = stmt.QueryRowContext(ctx, orderId).Scan(&order.OrderID, &order.UserID, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrOrderNotFound
		}
		return nil, fmt.Errorf("Postgres.GetOrderByID: exec query error: %w", err)
	}

	return &order, nil
}

func (s *Storage) GetOrdersByUserID(ctx context.Context, userId int64) ([]*model.Order, error) {
	stmt, err := s.db.PrepareContext(ctx, `SELECT o.id, o.number, s.status, s.accrual, o.uploaded_at
			FROM o orders JOIN s order_statuses ON o.status = s.id WHERE user_id = $1`)
	if err != nil {
		return nil, fmt.Errorf("Postgres.GetOrdersByUserID: prepare statement error: %w", err)
	}
	defer stmt.Close()
	var orders []*model.Order
	rows, err := stmt.QueryContext(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("Postgres.GetOrdersByUserID: exec query error: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var order model.Order
		order.UserID = userId
		err := rows.Scan(&order.OrderID, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, fmt.Errorf("Postgres.GetOrdersByUserID: rows.Scan error: %w", err)
		}
		orders = append(orders, &order)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("Postgres.GetOrdersByUserID: rows.Err error: %w", err)
	}
	return orders, nil
}
