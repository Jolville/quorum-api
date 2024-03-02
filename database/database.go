package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
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
		return "postgres://postgres:jesse@localhost:5432/quorum", nil
	}

	return "", fmt.Errorf("not implemented outside local env")
}

type UUIDSlice []uuid.UUID

func (u *UUIDSlice) Scan(src interface{}) error {
	switch src := src.(type) {
	case string:
		out := []uuid.UUID{}
		r := strings.NewReplacer("{", "", "}", "")
		rawUUIDs := strings.Split(r.Replace(src), ",")

		for _, rawUUID := range rawUUIDs {
			parsedUUID, err := uuid.Parse(rawUUID)
			if err != nil {
				return fmt.Errorf("parsing uuid %q: %w", rawUUID, err)
			}
			out = append(out, parsedUUID)
		}

		*u = out

	case nil:
		*u = []uuid.UUID{}
	default:
		return fmt.Errorf("unsupported type for UUIDSlice: %T", src)
	}

	return nil
}

func (u *UUIDSlice) Slice() []uuid.UUID {
	return []uuid.UUID(*u)
}

func (u UUIDSlice) Value() (driver.Value, error) {
	stringSlice := []string{}
	for _, elem := range u {
		stringSlice = append(stringSlice, elem.String())
	}
	return fmt.Sprintf("{%s}", strings.Join(stringSlice, ",")), nil
}

type EmailSlice []string

func (e EmailSlice) Value() (driver.Value, error) {
	return fmt.Sprintf("{%s}", strings.Join(e, ",")), nil
}
