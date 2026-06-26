package postgres

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/paulwwyvern/gophermart/internal/model"
	"github.com/paulwwyvern/gophermart/internal/model/errs"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorage_CreateOrder(t *testing.T) {
	userId := int64(1234)
	orderNumber := "123456"

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"user_id", "is_not_created"}).
		AddRow(int64(0), false)

	mock.ExpectPrepare(regexp.QuoteMeta("INSERT INTO orders (user_id, number)"))
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO orders (user_id, number)")).WithArgs(userId, orderNumber).WillReturnRows(rows)

	storage := &Storage{db: db}

	_, err = storage.CreateOrder(context.Background(), userId, orderNumber)
	assert.NoError(t, err)
}

func TestStorage_CreateOrder_Conflict(t *testing.T) {
	userId := int64(1234)
	orderNumber := "123456"

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"user_id", "is_not_created"}).
		AddRow(int64(5678), true)

	mock.ExpectPrepare(regexp.QuoteMeta("INSERT INTO orders (user_id, number)"))
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO orders (user_id, number)")).WithArgs(userId, orderNumber).WillReturnRows(rows)

	storage := &Storage{db: db}

	orderUserId, err := storage.CreateOrder(context.Background(), userId, orderNumber)

	assert.Equal(t, int64(5678), orderUserId)
	assert.ErrorIs(t, err, errs.ErrOrderAlreadyExists)
}

func TestStorage_GetOrderByID(t *testing.T) {
	orderId := int64(1234)
	userId := int64(5678)
	number := "123456"
	status := "NEW"
	accrual := decimal.NewFromFloat(3.14)
	uploadedAt := time.Now()

	order := &model.Order{
		OrderID:    orderId,
		UserID:     userId,
		Number:     number,
		Status:     status,
		Accrual:    accrual,
		UploadedAt: uploadedAt,
	}

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "user_id", "number", "status", "accrual", "uploaded_at"}).
		AddRow(orderId, userId, number, status, accrual, uploadedAt)

	query := "SELECT o.id, o.user_id, o.number, s.status, o.accrual, o.uploaded_at FROM orders"
	mock.ExpectPrepare(query)
	mock.ExpectQuery(query).WithArgs(orderId).WillReturnRows(rows)

	storage := &Storage{db: db}
	res, err := storage.GetOrderByID(context.Background(), orderId)
	assert.NoError(t, err)
	assert.Equal(t, order, res)
}

func TestStorage_GetOrderByID_NotFound(t *testing.T) {
	orderId := int64(1234)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "user_id", "number", "status", "accrual", "uploaded_at"})

	query := "SELECT o.id, o.user_id, o.number, s.status, o.accrual, o.uploaded_at FROM orders"
	mock.ExpectPrepare(query)
	mock.ExpectQuery(query).WithArgs(orderId).WillReturnRows(rows)

	storage := &Storage{db: db}
	_, err = storage.GetOrderByID(context.Background(), orderId)
	assert.Error(t, err, errs.ErrOrderNotFound)
}

func TestStorage_GetOrdersByUserID(t *testing.T) {

	now := time.Now()
	userId := int64(1234)

	orders := []*model.Order{
		{
			OrderID:    int64(1234),
			UserID:     userId,
			Number:     "123456",
			Status:     "NEW",
			Accrual:    decimal.NewFromFloat(3.14),
			UploadedAt: now,
		},
		{
			OrderID:    int64(5678),
			UserID:     userId,
			Number:     "654321",
			Status:     "PROCESSING",
			Accrual:    decimal.NewFromFloat(2.71),
			UploadedAt: now.Add(time.Hour),
		},
	}

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "number", "status", "accrual", "uploaded_at"}).
		AddRow(int64(1234), "123456", "NEW", decimal.NewFromFloat(3.14), now).
		AddRow(int64(5678), "654321", "PROCESSING", decimal.NewFromFloat(2.71), now.Add(time.Hour))

	query := "SELECT o.id, o.number, s.status, o.accrual, o.uploaded_at FROM orders"
	mock.ExpectPrepare(query)
	mock.ExpectQuery(query).WithArgs(userId).WillReturnRows(rows)

	storage := &Storage{db: db}

	res, err := storage.GetOrdersByUserID(context.Background(), userId)
	require.NoError(t, err)
	assert.Equal(t, orders, res)
}

func TestStorage_GetNewOrProcessingOrderNumbers(t *testing.T) {
	orders := []*model.ProcessingOrder{
		{
			UserID:      int64(1234),
			OrderNumber: "123456",
		},
		{
			UserID:      int64(5678),
			OrderNumber: "654321",
		},
	}

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"user_id", "order_number"}).
		AddRow(int64(1234), "123456").
		AddRow(int64(5678), "654321")

	query := "SELECT user_id, number FROM orders WHERE status = 0 OR status = 1"
	mock.ExpectPrepare(query)
	mock.ExpectQuery(query).WithArgs(2).WillReturnRows(rows)

	storage := &Storage{db: db}
	res, err := storage.GetNewOrProcessingOrderNumbers(context.Background(), 2)
	require.NoError(t, err)
	assert.Equal(t, orders, res)
}

func TestStorage_SetOrderStatus(t *testing.T) {
	number := "123456"
	status := "NEW"
	accrual := decimal.NewFromFloat(3.14)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	query := regexp.QuoteMeta("UPDATE orders SET status = $1, accrual = $2 WHERE number = $3")
	mock.ExpectPrepare(query)
	mock.ExpectExec(query).WithArgs(0, accrual, number).WillReturnResult(sqlmock.NewResult(1, 1))

	storage := &Storage{db: db}
	err = storage.SetOrderStatus(context.Background(), number, status, accrual)
	assert.NoError(t, err)

}
