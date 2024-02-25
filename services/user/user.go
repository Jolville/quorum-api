package srvuser

import (
	"context"
	"errors"
	"quorum-api/database"

	"github.com/google/uuid"
)

type SRVUser interface {
	GetUsersByFilter(ctx context.Context, request GetUsersByFilterRequest) ([]User, error)
	CreateUnverifiedUser(ctx context.Context, request CreateUnverifiedUserRequest) (uuid.UUID, error)
	VerifyUser(ctx context.Context, id uuid.UUID) error
}

type Profession string

const (
	Unspecified Profession = ""
	ProductDesigner Profession = "PRODUCT_DESIGNER"
	SoftwareEngineer Profession = "SOFTWARE_ENGINEER"
)

type GetUsersByFilterRequest struct {
	IDs []uuid.UUID
	Emails []string
}

type User struct {
	ID uuid.UUID
	Email string
	FirstName string
	LastName string
	Profression Profession
}

type CreateUnverifiedUserRequest struct {
	Email string
	FirstName string
	LastName string
	Profression Profession
}

var ErrEmailTaken = errors.New("another verified user exists with that email")
var ErrInvalidEmail = errors.New("email string format is invalid")
var ErrUserNotFound = errors.New("no verified user exists")

func New(db database.Q) SRVUser {
	return &srv{
		db: db,
	}
}

type srv struct {
	db database.Q
}

func (s *srv) GetUsersByFilter(ctx context.Context, request GetUsersByFilterRequest) ([]User, error) {
	panic("not implemented")
}

func (s *srv) CreateUnverifiedUser(ctx context.Context, request CreateUnverifiedUserRequest) (uuid.UUID, error) {
	panic("not implemented")
}

func (s *srv) VerifyUser(ctx context.Context, id uuid.UUID) error {
	panic("not implemented")
}

