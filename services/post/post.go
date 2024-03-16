package srvpost

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type SRVPost interface {
	GetPostsByFilter(ctx context.Context, request GetPostsByFilterRequest) ([]Post, error)
	CreatePost(ctx context.Context, request CreatePostRequest) error
	GetOptionsByFilter(ctx context.Context, request GetOptionsByFilterRequest) ([]Option, error)
	GetVotesByFilter(ctx context.Context, request GetVotesByFilterRequest) ([]Vote, error)
}

type GetPostsByFilterRequest struct {
	IDs []uuid.UUID
}

type Post struct {
	ID          uuid.UUID
	DesignPhase *DesignPhase
	Context     *string
	Category    *PostCategory
	LiveAt      *time.Time
	ClosesAt    *time.Time
	AuthorID    uuid.UUID
	Tags        []string
	OptionIDs   []uuid.UUID
	VoteIDs     []uuid.UUID
}

type Option struct {
	ID       uuid.UUID
	URL      *string
	Position int
}

type Vote struct {
	ID      uuid.UUID
	PostID  uuid.UUID
	VoterID uuid.UUID
	Reason  *string
}

type CreatePostRequest struct {
	ID          uuid.UUID
	DesignPhase *DesignPhase
	Context     *string
	Category    *PostCategory
	LiveAt      *time.Time
	ClosesAt    *time.Time
	Tags        []string
	Options     []*CreatePostOptionRequest
}

type Upload struct {
	File        io.ReadSeeker
	Filename    string
	Size        int64
	ContentType string
}

type CreatePostOptionRequest struct {
	Position int
	File     Upload
}

type GetOptionsByFilterRequest struct {
	IDs []uuid.UUID
}

type GetVotesByFilterRequest struct {
	IDs []uuid.UUID
}

var ErrTooManyOptions = errors.New("exceeded the maximum amount of options")

var ErrTooFewOptions = errors.New("at least 2 options are required to create a post")

var ErrPostNotOwned = errors.New("post with id already authored by another user")

var ErrLiveAtAlreadyPassed = errors.New("live at time has already passed")

var ErrNotOpenForLongEnough = errors.New("post must be open for at least 1 hour")

func New(db *sqlx.DB) SRVPost {
	return &srv{
		db: db,
	}
}

type srv struct {
	db *sqlx.DB
}

func (s *srv) GetPostsByFilter(
	ctx context.Context, request GetPostsByFilterRequest,
) ([]Post, error) {
	panic("not implemented")
}

func (s *srv) CreatePost(ctx context.Context, request CreatePostRequest) error {
	panic("not implemented")
}

func (s *srv) GetOptionsByFilter(
	ctx context.Context, request GetOptionsByFilterRequest,
) ([]Option, error) {
	panic("not implemented")
}

func (s *srv) GetVotesByFilter(
	ctx context.Context, request GetVotesByFilterRequest,
) ([]Vote, error) {
	panic("not implemented")
}
