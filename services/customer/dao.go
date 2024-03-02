package srvcustomer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"quorum-api/database"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/uuid"
)

type DBLock string

const (
	DBLockUnspecified DBLock = ""
	DBLockForUpdate   DBLock = "for update"
)

type getCustomersByFilterParams struct {
	IDs    database.UUIDSlice
	Emails []string
}

type customer struct {
	ID          uuid.UUID
	Email       string
	FirstName   string
	LastName    string
	Profression string
}

func getCustomersByFilter(
	ctx context.Context,
	db database.Q,
	params getCustomersByFilterParams,
	dbLock DBLock,
) ([]customer, error) {
	customers := []customer{}
	query := `
		select id, email, first_name, last_name, profression
		from verified_customer
		where deleted_at is null
	`

	args := []any{}
	if len(params.Emails) > 0 {
		args = append(args, params.Emails)
		query += " and email = any($1)"
	}
	if len(params.IDs) > 0 {
		args = append(args, params.IDs)
		query += " and id = any($1)"
	}

	query = fmt.Sprintf("%s %s", query, dbLock)

	spew.Dump(query)

	if err := db.SelectContext(ctx, &customers, query, args...); err != nil {
		return nil, fmt.Errorf("selecting customers: %w", err)
	}

	return customers, nil
}

type upsertUnverifiedCustomerParams struct {
	Email       string
	FirstName   string
	LastName    string
	Profression string
}

func upsertUnverifiedCustomer(
	ctx context.Context, q database.Q, params upsertUnverifiedCustomerParams,
) (uuid.UUID, error) {
	customerID := uuid.UUID{}
	if err := q.SelectContext(ctx, &customerID, `
		insert into unverified_customer (
			id, email, first_name, last_name, profression
		) values ($1, $2, $3, $4, $5)
		on conflict email do update set
			first_name = $3,
			last_name = $4,
			profression = $5,
			updated_at = now()
		returning id
		`,
		uuid.New(),
		params.Email,
		params.FirstName,
		params.LastName,
		params.Profression,
	); err != nil {
		return uuid.Nil, fmt.Errorf("inserting into unverified_customer: %w", err)
	}
	return customerID, nil
}

type upsertCustomerParams struct {
	Email       string
	FirstName   string
	LastName    string
	Profression string
}

func upsertCustomer(
	ctx context.Context, q database.Q, params upsertCustomerParams,
) (uuid.UUID, error) {
	customerID := uuid.UUID{}
	if err := q.SelectContext(ctx, &customerID, `
		insert into verified_customer (
			id, email, first_name, last_name, profression
		) values ($1, $2, $3, $4, $5)
		on conflict email do update set
			first_name = $3,
			last_name = $4,
			profression = $5,
			updated_at = now()
		returning id
		`,
		uuid.New(),
		params.Email,
		params.FirstName,
		params.LastName,
		params.Profression,
	); err != nil {
		return uuid.Nil, fmt.Errorf("inserting into unverified_customer: %w", err)
	}
	return customerID, nil
}

var errNoUnverifiedCustomer = errors.New("no unverified customer found")

func deleteUnverifiedCustomer(
	ctx context.Context, q database.Q, id uuid.UUID,
) (*customer, error) {
	customer := customer{}
	if err := q.GetContext(ctx, &customer, `
		update unverified_customer set deleted_at = now() where id = $1
		returning id, email, first_name, last_name, profression
	`, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, errNoUnverifiedCustomer
		}
		return nil, fmt.Errorf("updating unverified_customer: %w", err)
	}

	return &customer, nil
}
