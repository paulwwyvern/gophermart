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

type ctxKey string

var txKey ctxKey = "tx"

func insertTx(ctx context.Context, tx *sql.Tx) {
	ctx = context.WithValue(ctx, txKey, tx)
}

func getTx(ctx context.Context) *sql.Tx {
	if tx, ok := ctx.Value(txKey).(*sql.Tx); ok {
		return tx
	}
	return nil
}
