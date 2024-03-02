package srvcustomer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"quorum-api/database"

	"github.com/google/uuid"
)

type DBLock string

const (
	DBLockUnspecified DBLock = ""
	DBLockForUpdate   DBLock = "for update"
)

type getCustomersByFilterParams struct {
	IDs    []uuid.UUID
	Emails []string
}

type customer struct {
	ID         uuid.UUID `db:"id"`
	Email      string    `db:"email"`
	FirstName  string    `db:"first_name"`
	LastName   string    `db:"last_name"`
	Profession string    `db:"profession"`
}

func getCustomersByFilter(
	ctx context.Context,
	db database.Q,
	params getCustomersByFilterParams,
	dbLock DBLock,
) ([]customer, error) {
	customers := []customer{}
	query := `
		select id, email, first_name, last_name, profession
		from customer
		where deleted_at is null
	`

	args := []any{}
	if len(params.Emails) > 0 {
		args = append(args, params.Emails)
		query = fmt.Sprintf("%s and email = any($%v)", query, len(args))
	}
	if len(params.IDs) > 0 {
		args = append(args, params.IDs)
		query = fmt.Sprintf("%s and id = any($%v)", query, len(args))
	}

	query = fmt.Sprintf("%s %s", query, dbLock)

	if err := db.SelectContext(ctx, &customers, query, args...); err != nil {
		return nil, fmt.Errorf("selecting customers: %w", err)
	}

	return customers, nil
}

type upsertUnverifiedCustomerParams struct {
	Email      string
	FirstName  string
	LastName   string
	Profession string
}

func upsertUnverifiedCustomer(
	ctx context.Context, q database.Q, params upsertUnverifiedCustomerParams,
) (uuid.UUID, error) {
	customerID := uuid.UUID{}
	if err := q.GetContext(ctx, &customerID, `
		insert into unverified_customer (
			id, email, first_name, last_name, profession
		) values ($1, $2, $3, $4, $5)
		on conflict (email) do update set
			first_name = $3,
			last_name = $4,
			profession = $5
		returning id
		`,
		uuid.New(),
		params.Email,
		params.FirstName,
		params.LastName,
		params.Profession,
	); err != nil {
		return uuid.Nil, fmt.Errorf("inserting into unverified_customer: %w", err)
	}
	return customerID, nil
}

type upsertCustomerParams struct {
	ID         uuid.UUID
	Email      string
	FirstName  string
	LastName   string
	Profession string
}

func upsertCustomer(
	ctx context.Context, q database.Q, params upsertCustomerParams,
) error {
	if _, err := q.ExecContext(ctx, `
		insert into customer (
			id, email, first_name, last_name, profession
		) values ($1, $2, $3, $4, $5)
		on conflict (id) do update set
			first_name = $3,
			last_name = $4,
			profession = $5,
			updated_at = now()
		`,
		params.ID,
		params.Email,
		params.FirstName,
		params.LastName,
		params.Profession,
	); err != nil {
		return fmt.Errorf("inserting into customer: %w", err)
	}
	return nil
}

var errNoUnverifiedCustomer = errors.New("no unverified customer found")

func getUnverifiedCustomer(
	ctx context.Context, q database.Q, id uuid.UUID,
) (*customer, error) {
	customer := customer{}
	if err := q.GetContext(ctx, &customer, `
		select id, email, first_name, last_name, profession from unverified_customer where id = $1
	`, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, errNoUnverifiedCustomer
		}
		return nil, fmt.Errorf("updating unverified_customer: %w", err)
	}

	return &customer, nil
}
