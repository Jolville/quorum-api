package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
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
		if os.Getenv("DATABASE_URL") == "" {
			return "", fmt.Errorf("DATABASE_URL not set")
		}
		return os.Getenv("DATABASE_URL"), nil
	}
	mustGetenv := func(k string) string {
		v := os.Getenv(k)
		if v == "" {
			log.Fatalf("Warning: %s environment variable not set.", k)
		}
		return v
	}
	var (
		dbUser                 = mustGetenv("DB_IAM_USER")              // e.g. 'service-account-name@project-id.iam'
		dbName                 = mustGetenv("DB_NAME")                  // e.g. 'my-database'
		instanceConnectionName = mustGetenv("INSTANCE_CONNECTION_NAME") // e.g. 'project:region:instance'
		usePrivate             = mustGetenv("PRIVATE_IP")
	)

	d, err := cloudsqlconn.NewDialer(context.Background(), cloudsqlconn.WithIAMAuthN())
	if err != nil {
		return "", fmt.Errorf("cloudsqlconn.NewDialer: %w", err)
	}
	var opts []cloudsqlconn.DialOption
	if usePrivate != "" {
		opts = append(opts, cloudsqlconn.WithPrivateIP())
	}

	dsn := fmt.Sprintf("user=%s database=%s", dbUser, dbName)
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return "", err
	}

	config.DialFunc = func(ctx context.Context, network, instance string) (net.Conn, error) {
		return d.Dial(ctx, instanceConnectionName, opts...)
	}
	dbURI := stdlib.RegisterConnConfig(config)

	return dbURI, nil
}

type UUIDSlice []uuid.UUID

func (u *UUIDSlice) Scan(src interface{}) error {
	switch src := src.(type) {
	case string:
		if src == "NULL" {
			*u = []uuid.UUID{}
			return nil
		}
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
