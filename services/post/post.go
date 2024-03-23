package srvpost

import (
	"context"
	"errors"
	"fmt"
	"io"
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

type UpsertPostRequest struct {
	ID          uuid.UUID
	AuthorID    uuid.UUID
	DesignPhase *DesignPhase
	Context     *string
	Category    *PostCategory
	LiveAt      *time.Time
	ClosesAt    *time.Time
	Tags        []string
	Options     []*UpsertPostOptionRequest
}

type Upload struct {
	File        io.ReadSeeker
	Filename    string
	Size        int64
	ContentType string
}

type UpsertPostOptionRequest struct {
	Position int
	File     *Upload
	ID       uuid.UUID
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

var ErrFileTooLarge = errors.New("file must not be larger than 5MB")

var ErrUnsupportedFileType = errors.New("only PNG, JPG and GIF files are supported")

var ErrClosesAtNotSet = errors.New("must set close time when publishing a post")

var ErrOptionPositionsInvalid = errors.New("must provide unique position from 1 to 6")

var ErrOptionFileRequired = errors.New("file is required to create new post option")

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
	posts, err := getPostsByFilter(
		ctx, s.db, getPostsByFilterParams(request), DBLockUnspecified,
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
			LiveAt:      p.LiveAt,
			ClosesAt:    p.ClosesAt,
			AuthorID:    p.AuthorID,
			Tags:        p.Tags,
			OptionIDs:   p.OptionIDs,
			VoteIDs:     p.VoteIDs,
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

	postToUpsert := upsertPostParams{
		ID:          request.ID,
		AuthorID:    request.AuthorID,
		DesignPhase: request.DesignPhase,
		Context:     request.Context,
		Category:    request.Category,
		LiveAt:      request.LiveAt,
		ClosesAt:    request.ClosesAt,
	}
	if existingPost == nil {
		if postToUpsert.LiveAt != nil &&
			postToUpsert.LiveAt.Before(time.Now().Add(-time.Minute*10)) {
			return ErrLiveAtAlreadyPassed
		}

		if postToUpsert.LiveAt != nil &&
			postToUpsert.ClosesAt != nil &&
			postToUpsert.ClosesAt.Sub(*postToUpsert.LiveAt) < time.Hour {
			return ErrNotOpenForLongEnough
		}

		if len(request.Options) > 6 {
			return ErrTooManyOptions
		}

		postWillBeLive := postToUpsert.LiveAt != nil && postToUpsert.LiveAt.Before(time.Now())

		if postWillBeLive && len(request.Options) < 2 {
			return ErrTooFewOptions
		}

		if postWillBeLive && request.ClosesAt == nil {
			return ErrClosesAtNotSet
		}

		for _, o := range request.Options {
			if o.File == nil {
				return ErrOptionFileRequired
			}
			if o.File.Size > (5 << 20) {
				return ErrFileTooLarge
			}
			if o.File.ContentType != "image/png" &&
				o.File.ContentType != "image/jpeg" &&
				o.File.ContentType != "image/gif" {
				return ErrUnsupportedFileType
			}
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
			var fileRef string
			switch o.File.ContentType {
			case "image/jpeg":
				fileRef = fmt.Sprintf(
					// "https://storage.cloud.google.com/%s/post-options/%s.jpeg", here is the url for later
					"%s/post-options/%s.jpeg",
					s.bucketName, o.ID,
				)
			case "image/png":
				fileRef = fmt.Sprintf(
					"%s/post-options/%s.png",
					s.bucketName, o.ID,
				)
			case "image/gif":
				fileRef = fmt.Sprintf(
					"%s/post-options/%s.gif",
					s.bucketName, o.ID,
				)
			}
			optionsToInsert = append(optionsToInsert, postOption{
				ID:       o.ID,
				PostID:   postToUpsert.ID,
				Position: o.Position,
				FileRef:  fileRef,
			})
			wg.Add(1)
			go func(ref string, file Upload) {
				defer wg.Done()
				s.addObjToBucket(ctx, ref, file)
			}(fileRef, *o.File)
		}

		if err = insertPost(ctx, tx, postToUpsert); err != nil {
			return fmt.Errorf("inserting post: %w", err)
		}

		if len(optionsToInsert) > 0 {
			if err = insertPostOptions(ctx, tx, optionsToInsert); err != nil {
				return fmt.Errorf("inserting options: %w", err)
			}
		}

		tagsToInsert := []postTag{}
		for _, tag := range request.Tags {
			tagsToInsert = append(tagsToInsert, postTag{
				PostID: postToUpsert.ID,
				Tag:    tag,
			})
		}

		if len(tagsToInsert) > 0 {
			if err = insertPostTags(ctx, tx, tagsToInsert); err != nil {
				return fmt.Errorf("inserting tags: %w", err)
			}
		}

		wg.Wait()
		if err = tx.Commit(); err != nil {
			return fmt.Errorf("comitting tx: %w", err)
		}
		return nil
	}

	panic("update case not implemeted")
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

func (s *srv) addObjToBucket(ctx context.Context, fileRef string, file Upload) {
	obj := s.bucket.Object(fileRef)
	w := obj.NewWriter(ctx)
	if _, err := io.Copy(w, file.File); err != nil {
		panic(fmt.Sprintf("copying bytes: %v", err))
	}
	w.ContentType = file.ContentType
	w.CacheControl = "public,max-age=31536000" // one year
	if err := w.Close(); err != nil {
		panic(fmt.Sprintf("closing file: %v", err))
	}
}
