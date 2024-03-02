package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.44

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"quorum-api/graph/model"
	srvcustomer "quorum-api/services/customer"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// SignUp is the resolver for the signUp field.
func (r *mutationResolver) SignUp(ctx context.Context, input model.SignUpInput) (*model.SignUpPayload, error) {
	if strings.Contains(input.ReturnTo, ".") {
		return &model.SignUpPayload{
			Errors: []model.SignUpError{
				&model.InvalidReturnToError{
					Message: "Return to should not contain url scheme or host",
					Path:    []string{"input", "returnTo"},
				},
			},
		}, nil
	}
	customerID, err := r.Services.Customer.CreateUnverifiedCustomer(ctx,
		srvcustomer.CreateUnverifiedCustomerRequest{
			Email:     input.Email,
			FirstName: input.FirstName,
			LastName:  input.LastName,
		},
	)
	if err == srvcustomer.ErrEmailTaken {
		return &model.SignUpPayload{
			Errors: []model.SignUpError{
				&model.EmailTakenError{
					Message: "Email already in use.",
					Path:    []string{"input", "email"},
				},
			},
		}, nil
	}
	if err == srvcustomer.ErrInvalidEmail {
		return &model.SignUpPayload{
			Errors: []model.SignUpError{
				&model.EmailTakenError{
					Message: "Invalid format for email.",
					Path:    []string{"input", "email"},
				},
			},
		}, nil
	}
	if err != nil {
		log.Printf("error creating unverified customer: %v", err)
		return nil, fmt.Errorf("unexpected error occured")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, JWTClaims{
		IsVerified: false,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   customerID.String(),
		},
	})
	tokenString, err := token.SignedString([]byte(r.JWTSecret))
	if err != nil {
		log.Printf("error signing token: %v", err)
		return nil, fmt.Errorf("unexpected error occured")
	}
	queryParams := url.Values{}
	queryParams.Add("returnTo", input.ReturnTo)
	queryParams.Add("token", tokenString)
	magicLink := fmt.Sprintf("%s/verify?%s", os.Getenv("FRONTEND_URL"), queryParams.Encode())
	if os.Getenv("GO_ENV") == "local" {
		log.Printf("Magic link:\n%s", magicLink)
	} else {
		panic(fmt.Errorf("not implemented: GetLoginLink: sending email"))
	}
	return &model.SignUpPayload{}, nil
}

// GetLoginLink is the resolver for the getLoginLink field.
func (r *mutationResolver) GetLoginLink(ctx context.Context, input model.GetLoginLinkInput) (*model.GetLoginLinkPayload, error) {
	if strings.Contains(input.ReturnTo, ".") {
		return &model.GetLoginLinkPayload{
			Errors: []model.GetLoginLinkError{
				&model.InvalidReturnToError{
					Message: "Return to should not contain url scheme or host",
					Path:    []string{"input", "returnTo"},
				},
			},
		}, nil
	}

	customers, err := r.Services.Customer.GetCustomersByFilter(
		ctx, srvcustomer.GetCustomersByFilterRequest{
			Emails: []string{input.Email},
		},
	)
	if err != nil {
		panic(fmt.Errorf("getting customers: %v", err))
	}
	if len(customers) < 1 {
		return &model.GetLoginLinkPayload{
			Errors: []model.GetLoginLinkError{
				&model.CustomerNotFoundError{
					Message: "Customer not found, call signUp mutation",
				},
			},
		}, nil
	}

	customer := customers[0]
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, JWTClaims{
		IsVerified: false,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   customer.ID.String(),
		},
	})
	tokenString, err := token.SignedString([]byte(r.JWTSecret))
	if err != nil {
		log.Printf("error signing token: %v", err)
		return nil, fmt.Errorf("unexpected error occured")
	}
	queryParams := url.Values{}
	queryParams.Add("returnTo", input.ReturnTo)
	queryParams.Add("token", tokenString)
	magicLink := fmt.Sprintf("%s/verify?%s", os.Getenv("FRONTEND_URL"), queryParams.Encode())
	if os.Getenv("GO_ENV") == "local" {
		log.Printf("Magic link:\n%s", magicLink)
	} else {
		panic(fmt.Errorf("not implemented: GetLoginLink: sending email"))
	}
	return &model.GetLoginLinkPayload{}, nil
}

// VerifyCustomerToken is the resolver for the verifyCustomerToken field.
func (r *mutationResolver) VerifyCustomerToken(ctx context.Context, input model.VerifyCustomerTokenInput) (*model.VerifyCustomerTokenPayload, error) {
	token, err := jwt.ParseWithClaims(
		input.Token, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(r.JWTSecret), nil
		})
	if err != nil {
		if err == jwt.ErrTokenExpired {
			return &model.VerifyCustomerTokenPayload{
				Errors: []model.VerifyCustomerTokenError{
					&model.LinkExpiredError{
						Message: "Link has expired, either call sign up (new customers), or generate a new login link (existing customers)",
					},
				},
			}, nil
		}
		panic(fmt.Errorf("parsing token: %v", err))
	} else if claims, ok := token.Claims.(*JWTClaims); ok {
		if claims.IsVerified {
			panic("expected token to not be verified")
		}
		customerID, err := uuid.Parse(claims.Subject)
		if err != nil {
			panic("expected customerID to be a uuid")
		}
		r.Services.Customer.VerifyCustomer(ctx, customerID)
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, JWTClaims{
			IsVerified: true,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().AddDate(1, 0, 0)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				NotBefore: jwt.NewNumericDate(time.Now()),
				Subject:   customerID.String(),
			},
		})
		tokenString, err := token.SignedString([]byte(r.JWTSecret))
		if err != nil {
			log.Printf("error signing token: %v", err)
			return nil, fmt.Errorf("unexpected error occured")
		}
		customer, err := GetLoaders(ctx).CustomerLoader.Load(ctx, customerID)
		if err != nil {
			log.Printf("loading customer: %v", err)
			return nil, fmt.Errorf("unexpected error occured")
		}
		return &model.VerifyCustomerTokenPayload{
			NewToken: &tokenString,
			Customer: &customer,
		}, nil
	}
	panic(fmt.Errorf("unknown claims type, cannot proceed"))
}

// Customer is the resolver for the customer field.
func (r *queryResolver) Customer(ctx context.Context) (*srvcustomer.Customer, error) {
	verifiedCustomer := GetVerifiedCustomer(ctx)
	if !verifiedCustomer.Valid {
		return nil, nil
	}
	customer, err := GetLoaders(ctx).CustomerLoader.Load(ctx, verifiedCustomer.UUID)
	if err != nil {
		return nil, fmt.Errorf("getting customer: %w", err)
	}
	return &customer, nil
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
