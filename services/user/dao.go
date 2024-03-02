package srvuser

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

type getUsersByFilterParams struct {
	IDs    database.UUIDSlice
	Emails []string
}

type user struct {
	ID          uuid.UUID
	Email       string
	FirstName   string
	LastName    string
	Profression string
}

func getUsersByFilter(
	ctx context.Context,
	db database.Q,
	params getUsersByFilterParams,
	dbLock DBLock,
) ([]user, error) {
	users := []user{}
	query := `
		select id, email, first_name, last_name, profression
		from user
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

	query += fmt.Sprintf("%s %s", query, dbLock)

	if err := db.SelectContext(ctx, &users, query, args...); err != nil {
		return nil, fmt.Errorf("selecting users: %w", err)
	}

	return users, nil
}

type upsertUnverifiedUserParams struct {
	Email       string
	FirstName   string
	LastName    string
	Profression string
}

func upsertUnverifiedUser(
	ctx context.Context, q database.Q, params upsertUnverifiedUserParams,
) (uuid.UUID, error) {
	userID := uuid.UUID{}
	if err := q.SelectContext(ctx, &userID, `
		insert into unverified_user (
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
		return uuid.Nil, fmt.Errorf("inserting into unverified_user: %w", err)
	}
	return userID, nil
}

type upsertUserParams struct {
	Email       string
	FirstName   string
	LastName    string
	Profression string
}

func upsertUser(
	ctx context.Context, q database.Q, params upsertUserParams,
) (uuid.UUID, error) {
	userID := uuid.UUID{}
	if err := q.SelectContext(ctx, &userID, `
		insert into user (
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
		return uuid.Nil, fmt.Errorf("inserting into unverified_user: %w", err)
	}
	return userID, nil
}

var errNoUnverifiedUser = errors.New("no unverified user found")

func deleteUnverifiedUser(
	ctx context.Context, q database.Q, id uuid.UUID,
) (*user, error) {
	user := user{}
	if err := q.GetContext(ctx, &user, `
		update unverified_user set deleted_at = now() where id = $1
		returning id, email, first_name, last_name, profression
	`, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, errNoUnverifiedUser
		}
		return nil, fmt.Errorf("updating unverified_user: %w", err)
	}

	return &user, nil
}
