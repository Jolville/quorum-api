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
	srvuser "quorum-api/services/user"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
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
	userID, err := r.Services.User.CreateUnverifiedUser(ctx,
		srvuser.CreateUnverifiedUserRequest{
			Email:     input.Email,
			FirstName: input.FirstName,
			LastName:  input.LastName,
		},
	)
	if err == srvuser.ErrEmailTaken {
		return &model.SignUpPayload{
			Errors: []model.SignUpError{
				&model.EmailTakenError{
					Message: "Email already in use.",
					Path:    []string{"input", "email"},
				},
			},
		}, nil
	}
	if err == srvuser.ErrInvalidEmail {
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
		log.Printf("error creating unverified user: %v", err)
		return nil, fmt.Errorf("unexpected error occured")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"unverified_user_id": userID,
		"exp":                time.Now().Add(time.Hour).UTC(),
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

	users, err := r.Services.User.GetUsersByFilter(
		ctx, srvuser.GetUsersByFilterRequest{
			Emails: []string{input.Email},
		},
	)
	if err != nil {
		panic(fmt.Errorf("getting users: %v", err))
	}
	if len(users) < 1 {
		return &model.GetLoginLinkPayload{
			Errors: []model.GetLoginLinkError{
				&model.UserNotFoundError{
					Message: "User not found, call signUp mutation",
				},
			},
		}, nil
	}

	user := users[0]
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"unverified_user_id": user.ID,
		"exp":                time.Now().Add(time.Hour).UTC(),
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

// VerifyUserToken is the resolver for the verifyUserToken field.
func (r *mutationResolver) VerifyUserToken(ctx context.Context, input model.VerifyUserTokenInput) (*model.VerifyUserTokenPayload, error) {
	panic(fmt.Errorf("not implemented: VerifyUserToken - verifyUserToken"))
}

// User is the resolver for the user field.
func (r *queryResolver) User(ctx context.Context) (*srvuser.User, error) {
	return nil, nil
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
