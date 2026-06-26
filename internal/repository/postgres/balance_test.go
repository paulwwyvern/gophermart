package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/paulwwyvern/gophermart/internal/model"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorage_AddUserBalanceByID(t *testing.T) {
	userId := int64(1234)
	add := decimal.NewFromFloat(3.14)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectPrepare("UPDATE users SET balance")
	mock.ExpectExec("UPDATE users SET balance").WithArgs(add, userId).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectClose()

	storage := &Storage{db: db}

	err = storage.AddUserBalanceByID(context.Background(), userId, add)

	require.NoError(t, err)
}

func TestStorage_AddUserWithdrawnByID(t *testing.T) {
	userId := int64(1234)
	add := decimal.NewFromFloat(3.14)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectPrepare("UPDATE users SET withdrawn")
	mock.ExpectExec("UPDATE users SET withdrawn").WithArgs(add, userId).WillReturnResult(sqlmock.NewResult(1, 1))

	storage := &Storage{db: db}

	err = storage.AddUserWithdrawnByID(context.Background(), userId, add)

	require.NoError(t, err)
}

func TestStorage_GetUserBalanceByID(t *testing.T) {
	userId := int64(1234)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	mock.ExpectPrepare("SELECT balance, withdrawn FROM users")
	rows := sqlmock.NewRows([]string{"balance", "withdrawn"}).
		AddRow(decimal.NewFromFloat(3.14), decimal.NewFromFloat(2.71))
	mock.ExpectQuery("SELECT balance, withdrawn FROM users").WithArgs(userId).WillReturnRows(rows)

	storage := &Storage{db: db}
	balance, err := storage.GetUserBalanceByID(context.Background(), userId)
	require.NoError(t, err)
	require.Equal(t, decimal.NewFromFloat(3.14), balance.Balance)
	require.Equal(t, decimal.NewFromFloat(2.71), balance.Withdrawn)
}

func TestStorage_CreateWithdrawal(t *testing.T) {
	userId := int64(1234)
	orderNumber := "123456"
	sum := decimal.NewFromFloat(3.14)

	withdrawal := &model.Withdrawal{
		UserID:      userId,
		OrderNumber: orderNumber,
		Sum:         sum,
	}

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectPrepare("INSERT INTO withdrawals")
	mock.ExpectExec("INSERT INTO withdrawals").WithArgs(userId, orderNumber, sum).WillReturnResult(sqlmock.NewResult(1, 1))

	storage := &Storage{db: db}

	err = storage.CreateWithdrawal(context.Background(), withdrawal)
	assert.NoError(t, err)
}

func TestStorage_GetWithdrawalsByUserID(t *testing.T) {

	now := time.Now()
	userId := int64(1234)

	withdrawals := []*model.Withdrawal{
		{
			OrderNumber: "123456",
			Sum:         decimal.NewFromFloat(3.14),
			ProceedAt:   now,
		},
		{
			OrderNumber: "654321",
			Sum:         decimal.NewFromFloat(2.71),
			ProceedAt:   now.Add(time.Second),
		},
	}

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"order_number", "sum", "processed_at"}).
		AddRow("123456", decimal.NewFromFloat(3.14), now).
		AddRow("654321", decimal.NewFromFloat(2.71), now.Add(time.Second))

	mock.ExpectPrepare("SELECT order_number, sum, processed_at FROM withdrawals ")
	mock.ExpectQuery("SELECT order_number, sum, processed_at FROM withdrawals ").WithArgs(userId).WillReturnRows(rows)

	storage := &Storage{db: db}

	res, err := storage.GetWithdrawalsByUserID(context.Background(), userId)
	require.NoError(t, err)
	assert.Equal(t, res, withdrawals)
}
