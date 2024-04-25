package srvpost

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type SRVPost interface {
	GetPostsByFilter(ctx context.Context, request GetPostsByFilterRequest) ([]Post, error)
	UpsertPost(ctx context.Context, request UpsertPostRequest) error
	GetOptionsByFilter(ctx context.Context, request GetOptionsByFilterRequest) ([]Option, error)
	GetVotesByFilter(ctx context.Context, request GetVotesByFilterRequest) ([]Vote, error)
	GenerateSignedPostOptionURL(ctx context.Context, request GenerateSignedPostOptionURLRequest) (*GenerateSignedPostOptionURLResponse, error)
}

type GetPostsByFilterRequest struct {
	IDs []uuid.UUID
}

type Post struct {
	ID          uuid.UUID
	DesignPhase *DesignPhase
	Context     *string
	Category    *PostCategory
	OpensAt     *time.Time
	ClosesAt    *time.Time
	AuthorID    uuid.UUID
	OptionIDs   []uuid.UUID
	VoteIDs     []uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Option struct {
	ID       uuid.UUID
	URL      *string
	Position int
}

type Vote struct {
	ID         uuid.UUID
	PostID     uuid.UUID
	CustomerID uuid.UUID
	OptionID   uuid.UUID
	Reason     *string
}

type UpsertPostRequest struct {
	ID          uuid.UUID
	AuthorID    uuid.UUID
	DesignPhase *DesignPhase
	Context     *string
	Category    *PostCategory
	OpensAt     *time.Time
	ClosesAt    *time.Time
	Options     []*UpsertPostOptionRequest
}

type UpsertPostOptionRequest struct {
	Position   int
	BucketName string
	FileKey    string
	ID         uuid.UUID
}

type GetOptionsByFilterRequest struct {
	IDs []uuid.UUID
}

type GetVotesByFilterRequest struct {
	IDs []uuid.UUID
}

type GenerateSignedPostOptionURLRequest struct {
	FileName    string
	ContentType string
}

type GenerateSignedPostOptionURLResponse struct {
	BucketName string
	FileKey    string
	URL        string
}

var ErrTooManyOptions = errors.New("exceeded the maximum amount of options")

var ErrTooFewOptions = errors.New("at least 2 options are required to create a post")

var ErrPostNotOwned = errors.New("post with id already authored by another user")

var ErrOpensAtAlreadyPassed = errors.New("live at time has already passed")

var ErrOptionFileImmutable = errors.New("options file cannot be changed")

var ErrClosesAtNotAfterOpensAt = errors.New("post close time must be after open time")

var ErrClosesAtNotSet = errors.New("must set close time when publishing a post")

var ErrOptionPositionsInvalid = errors.New("must provide unique position from 1 to 6")

var ErrUnsupportedFileType = errors.New("only .jpeg, .png and .gif files are supported")

func New(db *sqlx.DB, bucket *storage.BucketHandle, bucketName string) SRVPost {
	return &srv{
		db:         db,
		bucket:     bucket,
		bucketName: bucketName,
	}
}

type srv struct {
	db         *sqlx.DB
	bucket     *storage.BucketHandle
	bucketName string
}

func (s *srv) GetPostsByFilter(
	ctx context.Context, request GetPostsByFilterRequest,
) ([]Post, error) {
	params := getPostsByFilterParams{
		IDs: request.IDs,
	}
	posts, err := getPostsByFilter(
		ctx, s.db, params, DBLockUnspecified,
	)
	if err != nil {
		return nil, fmt.Errorf("getting posts: %w", err)
	}

	res := []Post{}
	for _, p := range posts {
		res = append(res, Post{
			ID:          p.ID,
			DesignPhase: p.DesignPhase,
			Context:     p.Context,
			Category:    p.Category,
			OpensAt:     p.OpensAt,
			ClosesAt:    p.ClosesAt,
			AuthorID:    p.AuthorID,
			OptionIDs:   p.OptionIDs,
			VoteIDs:     p.VoteIDs,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		})
	}

	return res, nil
}

func (s *srv) UpsertPost(ctx context.Context, request UpsertPostRequest) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning tx: %w", err)
	}
	defer tx.Rollback()

	posts, err := getPostsByFilter(
		ctx, tx, getPostsByFilterParams{
			IDs: []uuid.UUID{request.ID},
		}, DBLockForUpdate,
	)
	if err != nil {
		return fmt.Errorf("getting posts: %w", err)
	}
	var existingPost *post
	if len(posts) == 1 {
		existingPost = &posts[0]
	}

	if existingPost == nil {
		postToUpsert := upsertPostParams{
			ID:          request.ID,
			AuthorID:    request.AuthorID,
			DesignPhase: request.DesignPhase,
			Context:     request.Context,
			Category:    request.Category,
			OpensAt:     request.OpensAt,
			ClosesAt:    request.ClosesAt,
		}
		if postToUpsert.OpensAt != nil &&
			postToUpsert.OpensAt.Before(time.Now().Add(-time.Minute*10)) {
			return ErrOpensAtAlreadyPassed
		}

		if postToUpsert.OpensAt != nil &&
			postToUpsert.ClosesAt != nil &&
			postToUpsert.ClosesAt.Before(*postToUpsert.OpensAt) {
			return ErrClosesAtNotAfterOpensAt
		}

		if len(request.Options) > 6 {
			return ErrTooManyOptions
		}

		postWillBeLive := postToUpsert.OpensAt != nil && postToUpsert.OpensAt.Before(time.Now())

		if postWillBeLive && len(request.Options) < 2 {
			return ErrTooFewOptions
		}

		if postWillBeLive && request.ClosesAt == nil {
			return ErrClosesAtNotSet
		}

		for i := 1; i <= len(request.Options); i++ {
			optionFound := false
			for _, o := range request.Options {
				if o.Position == i {
					optionFound = true
					break
				}
			}
			if !optionFound {
				return ErrOptionPositionsInvalid
			}
		}

		wg := sync.WaitGroup{}
		optionsToInsert := []postOption{}
		for _, o := range request.Options {
			if s.bucketName != o.BucketName {
				return fmt.Errorf("invalid bucket name")
			}
			fileRef := fmt.Sprintf(
				"%s/%s",
				s.bucketName, o.FileKey,
			)
			optionsToInsert = append(optionsToInsert, postOption{
				ID:       o.ID,
				PostID:   postToUpsert.ID,
				Position: o.Position,
				FileRef:  fileRef,
			})
			wg.Add(1)
			go func(fileKey string) {
				defer wg.Done()
				if _, err := s.bucket.Object(fileKey).If(storage.Conditions{
					DoesNotExist: true,
				}).Attrs(ctx); err != nil {
					if err == storage.ErrObjectNotExist {
						panic(fmt.Sprintf("file %s does not exist", fileKey))
					}
					panic(fmt.Errorf("verifying file exists: %v", err))
				}
			}(o.FileKey)
		}

		if err = upsertPost(ctx, tx, postToUpsert); err != nil {
			return fmt.Errorf("inserting post: %w", err)
		}

		if len(optionsToInsert) > 0 {
			if err = insertPostOptions(ctx, tx, optionsToInsert); err != nil {
				return fmt.Errorf("inserting options: %w", err)
			}
		}

		wg.Wait()
		if err = tx.Commit(); err != nil {
			return fmt.Errorf("comitting tx: %w", err)
		}
		return nil
	}

	postToUpsert := upsertPostParams{
		ID:          existingPost.ID,
		AuthorID:    existingPost.AuthorID,
		DesignPhase: existingPost.DesignPhase,
		Context:     existingPost.Context,
		Category:    existingPost.Category,
		OpensAt:     existingPost.OpensAt,
		ClosesAt:    existingPost.ClosesAt,
	}

	if request.AuthorID != existingPost.AuthorID {
		return ErrPostNotOwned
	}

	if !existingPost.OpensAt.After(time.Now()) {
		return ErrOpensAtAlreadyPassed
	}

	if postToUpsert.OpensAt != nil &&
		postToUpsert.OpensAt.Before(time.Now().Add(-time.Minute*10)) {
		return ErrOpensAtAlreadyPassed
	}

	if postToUpsert.OpensAt != nil &&
		postToUpsert.ClosesAt != nil &&
		postToUpsert.ClosesAt.Before(*postToUpsert.OpensAt) {
		return ErrClosesAtNotAfterOpensAt
	}

	postWillBeLive := postToUpsert.OpensAt != nil && postToUpsert.OpensAt.Before(time.Now())

	if postWillBeLive && request.ClosesAt == nil {
		return ErrClosesAtNotSet
	}

	if request.Category != nil {
		postToUpsert.Category = request.Category
	}
	if request.ClosesAt != nil {
		postToUpsert.ClosesAt = request.ClosesAt
	}
	if request.DesignPhase != nil {
		postToUpsert.DesignPhase = request.DesignPhase
	}
	if request.Context != nil {
		postToUpsert.Context = request.Context
	}
	if request.OpensAt != nil {
		postToUpsert.OpensAt = request.OpensAt
	}

	if postWillBeLive && len(request.Options) < 2 {
		return ErrTooFewOptions
	}

	if len(request.Options) > 6 {
		return ErrTooManyOptions
	}

	for i := 1; i <= len(request.Options); i++ {
		optionFound := false
		for _, o := range request.Options {
			if o.Position == i {
				optionFound = true
				break
			}
		}
		if !optionFound {
			return ErrOptionPositionsInvalid
		}
	}

	existingOptions, err := getPostOptionsByFilter(
		ctx, tx, getPostOptionsByFilterParams{
			IDs: existingPost.OptionIDs,
		}, DBLockForUpdate,
	)
	if err != nil {
		return fmt.Errorf("getting post options: %w", err)
	}

	optionIDsToDelete := []uuid.UUID{}
	optionsToUpdate := []updatePostOption{}
	for _, eo := range existingOptions {
		deleteExistingOption := true
		for _, no := range request.Options {
			if eo.ID == no.ID {
				deleteExistingOption = false
				if eo.Position != no.Position {
					optionsToUpdate = append(optionsToUpdate, updatePostOption{
						ID:       eo.ID,
						Position: no.Position,
					})
				}
			}
		}
		if deleteExistingOption {
			optionIDsToDelete = append(optionIDsToDelete, eo.ID)
		}
	}

	wg := sync.WaitGroup{}
	optionsToInsert := []postOption{}
	for _, o := range request.Options {
		if s.bucketName != o.BucketName {
			return fmt.Errorf("invalid bucket name")
		}
		fileRef := fmt.Sprintf(
			"%s/%s",
			s.bucketName, o.FileKey,
		)
		optionsToInsert = append(optionsToInsert, postOption{
			ID:       o.ID,
			PostID:   postToUpsert.ID,
			Position: o.Position,
			FileRef:  fileRef,
		})
		wg.Add(1)
		go func(fileKey string) {
			defer wg.Done()
			if _, err := s.bucket.Object(fileKey).If(storage.Conditions{
				DoesNotExist: true,
			}).Attrs(ctx); err != nil {
				if err == storage.ErrObjectNotExist {
					panic(fmt.Sprintf("file %s does not exist", fileKey))
				}
				panic(fmt.Errorf("verifying file exists: %v", err))
			}
		}(o.FileKey)
	}

	if err = upsertPost(ctx, tx, postToUpsert); err != nil {
		return fmt.Errorf("inserting post: %w", err)
	}

	if len(optionIDsToDelete) > 0 {
		// todo: probs want a cron to delete dangly files in bucket
		if err = deletePostOptions(ctx, tx, optionIDsToDelete); err != nil {
			return fmt.Errorf("deleting post options: %w", err)
		}
	}

	if len(optionsToUpdate) > 0 {
		if err = updatePostOptionPositions(ctx, tx, optionsToUpdate); err != nil {
			return fmt.Errorf("updating post option positions: %w", err)
		}
	}

	if len(optionsToInsert) > 0 {
		if err = insertPostOptions(ctx, tx, optionsToInsert); err != nil {
			return fmt.Errorf("inserting options: %w", err)
		}
	}

	wg.Wait()
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("comitting tx: %w", err)
	}
	return nil
}

func (s *srv) GetOptionsByFilter(
	ctx context.Context, request GetOptionsByFilterRequest,
) ([]Option, error) {
	params := getPostOptionsByFilterParams{
		IDs: request.IDs,
	}
	postOptions, err := getPostOptionsByFilter(
		ctx, s.db, params, DBLockUnspecified,
	)
	if err != nil {
		return nil, fmt.Errorf("getting posts: %w", err)
	}

	res := []Option{}
	for _, po := range postOptions {
		url := fmt.Sprintf("https://storage.cloud.google.com/%s/%s", s.bucketName, po.FileRef)
		res = append(res, Option{
			ID:       po.ID,
			Position: po.Position,
			URL:      &url,
		})
	}

	return res, nil
}

func (s *srv) GetVotesByFilter(
	ctx context.Context, request GetVotesByFilterRequest,
) ([]Vote, error) {
	params := getPostVotesByFilterParams{
		IDs: request.IDs,
	}
	postVotes, err := getPostVotesByFilter(
		ctx, s.db, params, DBLockUnspecified,
	)
	if err != nil {
		return nil, fmt.Errorf("getting posts: %w", err)
	}

	res := []Vote{}
	for _, pv := range postVotes {
		reason := *pv.Reason
		res = append(res, Vote{
			ID:         pv.ID,
			CustomerID: pv.CustomerID,
			OptionID:   pv.PostOptionID,
			PostID:     pv.PostID,
			Reason:     &reason,
		})
	}

	return res, nil
}

func (s *srv) GenerateSignedPostOptionURL(
	ctx context.Context,
	request GenerateSignedPostOptionURLRequest,
) (*GenerateSignedPostOptionURLResponse, error) {
	ext := filepath.Ext(strings.ToLower(request.FileName))
	if ext == "" {
		return nil, fmt.Errorf("expected file extension to be non-empty")
	}
	if ext != ".jpeg" && ext != ".png" && ext != ".gif" {
		return nil, ErrUnsupportedFileType
	}
	if request.ContentType == "" {
		return nil, fmt.Errorf("expected content type to be non-empty")
	}
	res := GenerateSignedPostOptionURLResponse{
		FileKey:    fmt.Sprintf("%s%s", uuid.NewString(), ext),
		BucketName: s.bucketName,
	}
	url, err := s.bucket.SignedURL(res.FileKey, &storage.SignedURLOptions{
		Method:      "PUT",
		Expires:     time.Now().Add(time.Minute * 15),
		ContentType: request.ContentType,
	})
	if err != nil {
		return nil, fmt.Errorf("creating SignedURL: %w", err)
	}
	res.URL = url
	return &res, nil
}
