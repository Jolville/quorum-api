package srvuser

import (
	"context"
	"errors"
	"fmt"
	"net/mail"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type SRVUser interface {
	GetUsersByFilter(ctx context.Context, request GetUsersByFilterRequest) ([]User, error)
	CreateUnverifiedUser(ctx context.Context, request CreateUnverifiedUserRequest) (uuid.UUID, error)
	VerifyUser(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
}

type GetUsersByFilterRequest struct {
	IDs    []uuid.UUID
	Emails []string
}

type User struct {
	ID          uuid.UUID
	Email       string
	FirstName   string
	LastName    string
	Profression string
}

type CreateUnverifiedUserRequest struct {
	Email       string
	FirstName   string
	LastName    string
	Profression string
}

var ErrEmailTaken = errors.New("another verified user exists with that email")
var ErrInvalidEmail = errors.New("email string format is invalid")
var ErrUserNotFound = errors.New("no verified user exists")

func New(db *sqlx.DB) SRVUser {
	return &srv{
		db: db,
	}
}

type srv struct {
	db *sqlx.DB
}

func (s *srv) GetUsersByFilter(ctx context.Context, request GetUsersByFilterRequest) ([]User, error) {
	users, err := getUsersByFilter(ctx, s.db, getUsersByFilterParams{
		IDs:    request.IDs,
		Emails: request.Emails,
	}, DBLockUnspecified)
	if err != nil {
		return nil, fmt.Errorf("getting users: %w", err)
	}

	res := []User{}
	for _, user := range users {
		res = append(res, User(user))
	}
	return res, nil
}

func (s *srv) CreateUnverifiedUser(ctx context.Context, request CreateUnverifiedUserRequest) (uuid.UUID, error) {
	mailAddress, err := mail.ParseAddress(request.Email)
	if err != nil {
		return uuid.Nil, ErrInvalidEmail
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return uuid.Nil, fmt.Errorf("beginning tx: %w", err)
	}
	defer tx.Rollback()

	users, err := getUsersByFilter(ctx, tx, getUsersByFilterParams{
		Emails: []string{mailAddress.Address},
	}, DBLockForUpdate)
	if err != nil {
		return uuid.Nil, fmt.Errorf("getting users: %w", err)
	}
	if len(users) > 0 {
		return uuid.Nil, ErrEmailTaken
	}

	userID, err := upsertUnverifiedUser(ctx, tx, upsertUnverifiedUserParams(request))
	if err != nil {
		return uuid.Nil, fmt.Errorf("upserting user: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return uuid.Nil, fmt.Errorf("committing tx: %w", err)
	}
	return userID, nil
}

func (s *srv) VerifyUser(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return uuid.Nil, fmt.Errorf("beginning tx: %w", err)
	}
	defer tx.Rollback()

	users, err := getUsersByFilter(ctx, tx, getUsersByFilterParams{
		IDs: []uuid.UUID{id},
	}, DBLockForUpdate)
	if err != nil {
		return uuid.Nil, fmt.Errorf("getting users: %w", err)
	}
	if len(users) > 0 {
		return id, nil
	}

	user, err := deleteUnverifiedUser(ctx, tx, id)
	if err != nil {
		if err == errNoUnverifiedUser {
			return uuid.Nil, ErrUserNotFound
		}
		return uuid.Nil, fmt.Errorf("deleting unverified user: %w", err)
	}

	userID, err := upsertUser(ctx, tx, upsertUserParams{
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Profression: user.Profression,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("upserting user: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return uuid.Nil, fmt.Errorf("committing tx: %w", err)
	}

	return userID, nil
}
