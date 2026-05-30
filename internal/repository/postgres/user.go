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

func (s *Storage) CreateUser(ctx context.Context, user *model.User) error {
	stmt, err := s.db.PrepareContext(ctx, `INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id`)
	if err != nil {
		return fmt.Errorf("Postgres.CreateUser: prepare statement error: %w", err)
	}
	defer stmt.Close()

	var userId int64
	err = stmt.QueryRowContext(ctx, user.Login, user.Password).Scan(&userId)
	if err != nil {
		var pgxErr *pgconn.PgError
		if errors.As(err, &pgxErr) {
			if pgxErr.Code == pgerrcode.UniqueViolation {
				return errs.ErrUserAlreadyExists
			}
		}
		return fmt.Errorf("CreateUser: query error: %w", err)
	}

	user.UserID = userId

	return nil
}

func (s *Storage) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	stmt, err := s.db.PrepareContext(ctx, `SELECT id, login, password_hash FROM users WHERE login = $1`)
	if err != nil {
		return nil, fmt.Errorf("Postgres.GetUserByLogin: prepare statement error: %w", err)
	}
	defer stmt.Close()

	user := &model.User{}
	err = stmt.QueryRowContext(ctx, login).Scan(&user.UserID, &user.Login, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrUserNotFound
		}
		return nil, fmt.Errorf("GetUserByLogin: query error: %w", err)
	}

	return user, nil
}
