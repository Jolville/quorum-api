package srvcustomer

import (
	"context"
	"errors"
	"fmt"
	"net/mail"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type SRVCustomer interface {
	GetCustomersByFilter(ctx context.Context, request GetCustomersByFilterRequest) ([]Customer, error)
	CreateUnverifiedCustomer(ctx context.Context, request CreateUnverifiedCustomerRequest) (uuid.UUID, error)
	VerifyCustomer(ctx context.Context, id uuid.UUID) error
}

type GetCustomersByFilterRequest struct {
	IDs    []uuid.UUID
	Emails []string
}

type Customer struct {
	ID         uuid.UUID
	Email      string
	FirstName  string
	LastName   string
	Profession string
}

type CreateUnverifiedCustomerRequest struct {
	Email      string
	FirstName  string
	LastName   string
	Profession string
}

var ErrEmailTaken = errors.New("another verified customer exists with that email")
var ErrInvalidEmail = errors.New("email string format is invalid")
var ErrCustomerNotFound = errors.New("no customer exists")

func New(db *sqlx.DB) SRVCustomer {
	return &srv{
		db: db,
	}
}

type srv struct {
	db *sqlx.DB
}

func (s *srv) GetCustomersByFilter(ctx context.Context, request GetCustomersByFilterRequest) ([]Customer, error) {
	customers, err := getCustomersByFilter(ctx, s.db, getCustomersByFilterParams{
		IDs:    request.IDs,
		Emails: request.Emails,
	}, DBLockUnspecified)
	if err != nil {
		return nil, fmt.Errorf("getting customers: %w", err)
	}

	res := []Customer{}
	for _, customer := range customers {
		res = append(res, Customer(customer))
	}
	return res, nil
}

func (s *srv) CreateUnverifiedCustomer(ctx context.Context, request CreateUnverifiedCustomerRequest) (uuid.UUID, error) {
	mailAddress, err := mail.ParseAddress(request.Email)
	if err != nil {
		return uuid.Nil, ErrInvalidEmail
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return uuid.Nil, fmt.Errorf("beginning tx: %w", err)
	}
	defer tx.Rollback()

	customers, err := getCustomersByFilter(ctx, tx, getCustomersByFilterParams{
		Emails: []string{mailAddress.Address},
	}, DBLockForUpdate)
	if err != nil {
		return uuid.Nil, fmt.Errorf("getting customers: %w", err)
	}
	if len(customers) > 0 {
		return uuid.Nil, ErrEmailTaken
	}

	customerID, err := upsertUnverifiedCustomer(ctx, tx, upsertUnverifiedCustomerParams(request))
	if err != nil {
		return uuid.Nil, fmt.Errorf("upserting customer: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return uuid.Nil, fmt.Errorf("committing tx: %w", err)
	}
	return customerID, nil
}

func (s *srv) VerifyCustomer(ctx context.Context, id uuid.UUID) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning tx: %w", err)
	}
	defer tx.Rollback()

	customers, err := getCustomersByFilter(ctx, tx, getCustomersByFilterParams{
		IDs: []uuid.UUID{id},
	}, DBLockForUpdate)
	if err != nil {
		return fmt.Errorf("getting customers: %w", err)
	}
	if len(customers) > 0 {
		return nil
	}

	customer, err := getUnverifiedCustomer(ctx, tx, id)
	if err != nil {
		if err == errNoUnverifiedCustomer {
			return ErrCustomerNotFound
		}
		return fmt.Errorf("deleting unverified customer: %w", err)
	}

	if err = upsertCustomer(ctx, tx, upsertCustomerParams{
		ID:         customer.ID,
		Email:      customer.Email,
		FirstName:  customer.FirstName,
		LastName:   customer.LastName,
		Profession: customer.Profession,
	}); err != nil {
		return fmt.Errorf("upserting customer: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("committing tx: %w", err)
	}

	return nil
}
