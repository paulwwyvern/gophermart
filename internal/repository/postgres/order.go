package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/paulwwyvern/gophermart/internal/model"
	"github.com/paulwwyvern/gophermart/internal/model/errs"
	"github.com/shopspring/decimal"
)

var orderStatuses = map[string]int{
	"NEW":        0,
	"PROCESSING": 1,
	"INVALID":    2,
	"PROCESSED":  3,
}

func (s *Storage) CreateOrder(ctx context.Context, userId int64, number string) (int64, error) {
	query := "INSERT INTO orders (user_id, number) " +
		"VALUES ($1, $2) ON CONFLICT(number) DO UPDATE SET number = excluded.number " +
		"RETURNING user_id, (xmax != 0);"

	stmt, err := s.Prepare(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("Postgres.CreateOrder: prepare statement error: %w", err)
	}
	defer stmt.Close()

	var orderUserId int64
	var isNotCreateOrder bool
	err = stmt.QueryRowContext(ctx, userId, number).Scan(&orderUserId, &isNotCreateOrder)
	if err != nil {
		return 0, fmt.Errorf("Postgres.CreateOrder: exec query error: %w", err)
	}
	if isNotCreateOrder {
		return orderUserId, errs.ErrOrderAlreadyExists
	}

	return 0, nil
}

func (s *Storage) GetOrderByID(ctx context.Context, orderId int64) (*model.Order, error) {
	query := "SELECT o.id, o.user_id, o.number, s.status, o.accrual, o.uploaded_at " +
		"FROM orders o JOIN order_statuses s ON o.status = s.id WHERE order_id = $1"

	stmt, err := s.Prepare(ctx, query)
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
	query := "SELECT o.id, o.number, s.status, o.accrual, o.uploaded_at " +
		"FROM orders o JOIN order_statuses s ON o.status = s.id WHERE user_id = $1 ORDER BY o.uploaded_at DESC"

	stmt, err := s.Prepare(ctx, query)
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

func (s *Storage) GetNewOrProcessingOrderNumbers(ctx context.Context, batchSize int) ([]*model.ProcessingOrder, error) {
	query := "SELECT user_id, number FROM orders WHERE status = 0 OR status = 1 ORDER BY updated_at ASC LIMIT $1 FOR UPDATE SKIP LOCKED"

	stmt, err := s.Prepare(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("Postgres.GetNewOrProcessingOrderNumbers: prepare statement error: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, batchSize)
	if err != nil {
		return nil, fmt.Errorf("Postgres.GetNewOrProcessingOrderNumbers: exec query error: %w", err)
	}
	defer rows.Close()

	var processingOrders []*model.ProcessingOrder
	for rows.Next() {
		processingOrder := &model.ProcessingOrder{}
		err = rows.Scan(&processingOrder.UserID, &processingOrder.OrderNumber)
		if err != nil {
			return nil, fmt.Errorf("Postgres.GetNewOrProcessingOrderNumbers: rows.Scan error: %w", err)
		}
		processingOrders = append(processingOrders, processingOrder)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("Postgres.GetNewOrProcessingOrderNumbers: rows.Err error: %w", err)
	}
	return processingOrders, nil
}

func (s *Storage) SetOrderStatus(ctx context.Context, number string, status string, accrual decimal.Decimal) error {
	query := "UPDATE orders SET status = $1, accrual = $2 WHERE number = $3"

	stmt, err := s.Prepare(ctx, query)
	if err != nil {
		return fmt.Errorf("Postgres.SetOrderStatus: prepare statement error: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, orderStatuses[status], accrual, number)
	if err != nil {
		return fmt.Errorf("Postgres.SetOrderStatus: exec query error: %w", err)
	}
	return nil
}
