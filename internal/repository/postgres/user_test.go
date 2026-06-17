package postgres

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/paulwwyvern/gophermart/internal/model"
	"github.com/paulwwyvern/gophermart/internal/model/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorage_CreateUser(t *testing.T) {
	userId := int64(1234)
	login := "abcde"
	password := "123456"

	user := &model.User{
		Login:    login,
		Password: password,
	}

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"user_id"}).
		AddRow(userId)

	query := regexp.QuoteMeta("INSERT INTO users ")
	mock.ExpectPrepare(regexp.QuoteMeta(query))
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(login, password).WillReturnRows(rows)

	storage := &Storage{db: db}

	ret, err := storage.CreateUser(context.Background(), user)
	assert.NoError(t, err)
	assert.Equal(t, userId, ret)
}

func TestStorage_CreateUser_Conflict(t *testing.T) {
	login := "abcde"
	password := "123456"

	user := &model.User{
		Login:    login,
		Password: password,
	}

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	retErr := &pgconn.PgError{
		Code: pgerrcode.UniqueViolation,
	}

	query := regexp.QuoteMeta("INSERT INTO users ")
	mock.ExpectPrepare(regexp.QuoteMeta(query))
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(login, password).WillReturnError(retErr)

	storage := &Storage{db: db}

	_, err = storage.CreateUser(context.Background(), user)
	assert.ErrorIs(t, err, errs.ErrUserAlreadyExists)
}

func TestStorage_GetUserByLogin(t *testing.T) {
	userId := int64(1234)

	login := "abcde"
	password := "123456"

	user := &model.User{
		UserID:   userId,
		Login:    login,
		Password: password,
	}

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "login", "password_hash"}).
		AddRow(userId, login, password)

	query := "SELECT id, login, password_hash FROM users"
	mock.ExpectPrepare(query)
	mock.ExpectQuery(query).WithArgs(login).WillReturnRows(rows)

	storage := &Storage{db: db}
	res, err := storage.GetUserByLogin(context.Background(), login)
	assert.NoError(t, err)
	assert.Equal(t, user, res)
}
