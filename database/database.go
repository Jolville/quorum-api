package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
)

type Q interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

func New(uri string) (*sqlx.DB, error) {
	driver := "pgx"

	db, err := sqlx.Connect(driver, uri)
	if err != nil {
		return nil, fmt.Errorf("connecting: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(time.Minute * 5)

	return db, nil
}

func GetConnectionStringFromEnv() (string, error) {
	if os.Getenv("GO_ENV") == "local" {
		return "postgres:jesse@localhost:5432/quorum?ssl_mode=disable", nil
	}

	return "", fmt.Errorf("not implemented outside local env")
}