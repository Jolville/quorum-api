package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.44

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"quorum-api/graph/model"
	srvcustomer "quorum-api/services/customer"
	srvpost "quorum-api/services/post"
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
			Email:      input.Email,
			FirstName:  &input.FirstName,
			LastName:   &input.LastName,
			Profession: &input.Profession,
		},
	)
	switch err {
	case nil:
		// continue
	case srvcustomer.ErrEmailTaken:
		customers, err := r.Services.Customer.GetCustomersByFilter(ctx,
			srvcustomer.GetCustomersByFilterRequest{
				Emails: []string{strings.ToLower(input.Email)},
			},
		)
		if err != nil {
			log.Printf("error getting customers: %v", err)
			return nil, fmt.Errorf("unexpected error occured")
		}
		if len(customers) != 1 {
			log.Printf("expected 1 customer")
			return nil, fmt.Errorf("unexpected error occured")
		}
		customerID = customers[0].ID
	case srvcustomer.ErrInvalidEmail:
		return &model.SignUpPayload{
			Errors: []model.SignUpError{
				&model.InvalidEmailError{
					Message: "Invalid format for email.",
					Path:    []string{"input", "email"},
				},
			},
		}, nil
	default:
		log.Printf("error creating unverified customer: %v", err)
		return nil, fmt.Errorf("unexpected error occured")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, JWTClaims{
		IsVerified: false,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
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
		return nil, fmt.Errorf("not implemented: GetLoginLink: sending email")
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
		return nil, fmt.Errorf("getting customers: %v", err)
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
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
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
		return nil, fmt.Errorf("not implemented: GetLoginLink: sending email")
	}
	return &model.GetLoginLinkPayload{}, nil
}

// VerifyCustomerToken is the resolver for the verifyCustomerToken field.
func (r *mutationResolver) VerifyCustomerToken(ctx context.Context, input model.VerifyCustomerTokenInput) (*model.VerifyCustomerTokenPayload, error) {
	token, err := jwt.ParseWithClaims(
		input.Token, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(r.JWTSecret), nil
		})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return &model.VerifyCustomerTokenPayload{
				Errors: []model.VerifyCustomerTokenError{
					&model.LinkExpiredError{
						Message: "Link has expired, either call sign up (new customers), or generate a new login link (existing customers)",
					},
				},
			}, nil
		}
		return nil, fmt.Errorf("parsing token: %v", err)
	} else if claims, ok := token.Claims.(*JWTClaims); ok {
		if claims.IsVerified {
			return nil, fmt.Errorf("expected token to not be verified")
		}
		customerID, err := uuid.Parse(claims.Subject)
		if err != nil {
			return nil, fmt.Errorf("expected customerID to be a uuid")
		}
		if err = r.Services.Customer.VerifyCustomer(ctx, customerID); err != nil {
			return nil, fmt.Errorf("verifying customer: %v", err)
		}
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
			Customer: customer,
		}, nil
	}
	return nil, fmt.Errorf("unknown claims type, cannot proceed")
}

// UpsertPost is the resolver for the upsertPost field.
func (r *mutationResolver) UpsertPost(ctx context.Context, input model.UpsertPostInput) (*model.UpsertPostPayload, error) {
	verifiedCustomer := GetVerifiedCustomer(ctx)
	if !verifiedCustomer.Valid {
		return &model.UpsertPostPayload{
			Errors: []model.UpsertPostError{
				model.AuthorUnknownError{
					Message: "Author of post unknown - try logging again",
				},
			},
		}, nil
	}

	options := []*srvpost.UpsertPostOptionRequest{}
	for _, o := range input.Options {
		var file *srvpost.Upload
		if o.File != nil {
			file = &srvpost.Upload{
				File:        o.File.File,
				Size:        o.File.Size,
				Filename:    o.File.Filename,
				ContentType: o.File.ContentType,
			}
		}
		options = append(options, &srvpost.UpsertPostOptionRequest{
			Position: o.Position,
			ID:       o.ID,
			File:     file,
		})
	}

	err := r.Services.Post.UpsertPost(ctx, srvpost.UpsertPostRequest{
		ID:          input.ID,
		Options:     options,
		DesignPhase: (*srvpost.DesignPhase)(input.DesignPhase),
		Category:    (*srvpost.PostCategory)(input.Category),
		OpensAt:     input.OpensAt,
		ClosesAt:    input.ClosesAt,
		Tags:        input.Tags,
		AuthorID:    verifiedCustomer.UUID,
	})
	if errors.Is(err, srvpost.ErrOpensAtAlreadyPassed) {
		return &model.UpsertPostPayload{
			Errors: []model.UpsertPostError{
				model.OpensAtAlreadyPassedError{
					Message: err.Error(),
				},
			},
		}, nil
	}
	if errors.Is(err, srvpost.ErrPostNotOwned) {
		return &model.UpsertPostPayload{
			Errors: []model.UpsertPostError{
				model.ErrPostNotOwned{
					Message: err.Error(),
				},
			},
		}, nil
	}
	if errors.Is(err, srvpost.ErrTooFewOptions) {
		return &model.UpsertPostPayload{
			Errors: []model.UpsertPostError{
				model.TooFewOptionsError{
					Message: err.Error(),
				},
			},
		}, nil
	}
	if errors.Is(err, srvpost.ErrTooManyOptions) {
		return &model.UpsertPostPayload{
			Errors: []model.UpsertPostError{
				model.TooManyOptionsError{
					Message: err.Error(),
				},
			},
		}, nil
	}
	if errors.Is(err, srvpost.ErrFileTooLarge) {
		return &model.UpsertPostPayload{
			Errors: []model.UpsertPostError{
				model.FileTooLargeError{
					Message: err.Error(),
				},
			},
		}, nil
	}
	if errors.Is(err, srvpost.ErrUnsupportedFileType) {
		return &model.UpsertPostPayload{
			Errors: []model.UpsertPostError{
				model.UnsupportedFileTypeError{
					Message: err.Error(),
				},
			},
		}, nil
	}
	// todo handle the other known errors
	if err != nil {
		return nil, fmt.Errorf("creating post: %w", err)
	}

	post, err := GetLoaders(ctx).PostLoader.Load(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("loading post: %w", err)
	}

	return &model.UpsertPostPayload{
		Post:   post,
		Errors: []model.UpsertPostError{},
	}, nil
}

// DesignPhase is the resolver for the designPhase field.
func (r *postResolver) DesignPhase(ctx context.Context, obj *srvpost.Post) (*model.DesignPhase, error) {
	return (*model.DesignPhase)(obj.DesignPhase), nil
}

// Category is the resolver for the category field.
func (r *postResolver) Category(ctx context.Context, obj *srvpost.Post) (*model.PostCategory, error) {
	return (*model.PostCategory)(obj.Category), nil
}

// Author is the resolver for the author field.
func (r *postResolver) Author(ctx context.Context, obj *srvpost.Post) (*srvcustomer.Customer, error) {
	customer, err := GetLoaders(ctx).CustomerLoader.Load(ctx, obj.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("loading author: %w", err)
	}
	return customer, nil
}

// Options is the resolver for the options field.
func (r *postResolver) Options(ctx context.Context, obj *srvpost.Post) ([]*srvpost.Option, error) {
	options, err := GetLoaders(ctx).PostOptionLoader.LoadAll(ctx, obj.OptionIDs)
	if err != nil {
		return nil, fmt.Errorf("loading options: %w", err)
	}
	return options, nil
}

// Votes is the resolver for the votes field.
func (r *postResolver) Votes(ctx context.Context, obj *srvpost.Post) ([]*srvpost.Vote, error) {
	votes, err := GetLoaders(ctx).PostVoteLoader.LoadAll(ctx, obj.VoteIDs)
	if err != nil {
		return nil, fmt.Errorf("loading votes: %w", err)
	}
	return votes, nil
}

// Status is the resolver for the status field.
func (r *postResolver) Status(ctx context.Context, obj *srvpost.Post) (model.PostStatus, error) {
	if obj == nil || obj.OpensAt == nil || obj.ClosesAt == nil {
		return model.PostStatusDraft, nil
	}
	if obj.OpensAt.After(time.Now()) {
		return model.PostStatusDraft, nil
	}
	if obj.ClosesAt.After(time.Now()) {
		return model.PostStatusLive, nil
	}
	return model.PostStatusClosed, nil
}

// Post is the resolver for the post field.
func (r *postVoteResolver) Post(ctx context.Context, obj *srvpost.Vote) (*srvpost.Post, error) {
	post, err := GetLoaders(ctx).PostLoader.Load(ctx, obj.PostID)
	if err != nil {
		return nil, fmt.Errorf("loading post: %w", err)
	}
	return post, nil
}

// Voter is the resolver for the voter field.
func (r *postVoteResolver) Voter(ctx context.Context, obj *srvpost.Vote) (*srvcustomer.Customer, error) {
	customer, err := GetLoaders(ctx).CustomerLoader.Load(ctx, obj.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("loading author: %w", err)
	}
	return customer, nil
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
	return customer, nil
}

// Post is the resolver for the post field.
func (r *queryResolver) Post(ctx context.Context, id uuid.UUID) (*srvpost.Post, error) {
	post, err := GetLoaders(ctx).PostLoader.Load(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("loading post: %w", err)
	}
	return post, nil
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Post returns PostResolver implementation.
func (r *Resolver) Post() PostResolver { return &postResolver{r} }

// PostVote returns PostVoteResolver implementation.
func (r *Resolver) PostVote() PostVoteResolver { return &postVoteResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type postResolver struct{ *Resolver }
type postVoteResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
