package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(databaseUrl string) (*Storage, error) {
	db, err := sql.Open("pgx", databaseUrl)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

func Migrate(source string, databaseDsn string) error {
	databaseDsn = strings.TrimPrefix(databaseDsn, "postgres://")
	databaseDsn = strings.TrimPrefix(databaseDsn, "postgresql://")

	databaseDsn = "pgx5://" + databaseDsn

	source = "file://" + source

	m, err := migrate.New(source, databaseDsn)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func (s *Storage) Close() {
	s.db.Close()
}

func (s *Storage) Prepare(ctx context.Context, query string) (*sql.Stmt, error) {
	var stmt *sql.Stmt
	var err error
	if tx := getTx(ctx); tx != nil {
		stmt, err = tx.PrepareContext(ctx, query)
	} else {
		stmt, err = s.db.PrepareContext(ctx, query)
	}
	return stmt, err
}

func (s *Storage) BeginTx(ctx context.Context) (context.Context, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return insertTx(ctx, tx), nil
}

func (s *Storage) CommitTx(ctx context.Context) error {
	tx := getTx(ctx)
	if tx == nil {
		return errors.New("Commit: no transaction found")
	}
	return tx.Commit()
}

func (s *Storage) RollbackTx(ctx context.Context) error {
	tx := getTx(ctx)
	if tx == nil {
		return errors.New("Rollback: no transaction found")
	}
	return tx.Rollback()
}

type ctxKey string

var txKey ctxKey = "tx"

func insertTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

func getTx(ctx context.Context) *sql.Tx {
	if tx, ok := ctx.Value(txKey).(*sql.Tx); ok {
		return tx
	}
	return nil
}
