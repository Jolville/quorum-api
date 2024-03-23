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
	OpensAt     *time.Time
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
	OpensAt     *time.Time
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

var ErrOpensAtAlreadyPassed = errors.New("live at time has already passed")

var ErrOptionFileImmutable = errors.New("options file cannot be changed")

var ErrClosesAtNotAfterOpensAt = errors.New("post close time must be after open time")

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
			OpensAt:     p.OpensAt,
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

		if err = upsertPost(ctx, tx, postToUpsert); err != nil {
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
			if err = upsertPostTags(ctx, tx, tagsToInsert); err != nil {
				return fmt.Errorf("inserting tags: %w", err)
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
		return fmt.Errorf("getting post options: %w")
	}

	optionIDsToDelete := []uuid.UUID{}
	optionsToUpdate := []updatePostOption{}
	for _, eo := range existingOptions {
		deleteExistingOption := true
		for _, no := range request.Options {
			if eo.ID == no.ID {
				deleteExistingOption = false
				if no.File != nil {
					return ErrOptionFileImmutable
				}
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
		if o.File == nil {
			existingOptionFound := false
			for _, eo := range existingOptions {
				if eo.ID == o.ID {
					existingOptionFound = true
					break
				}
			}
			if !existingOptionFound {
				return ErrOptionFileRequired
			}
		}

		if o.File.Size > (5 << 20) {
			return ErrFileTooLarge
		}

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
		default:
			return ErrUnsupportedFileType
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

	tagsToUpsert := []postTag{}
	for _, tag := range request.Tags {
		tagsToUpsert = append(tagsToUpsert, postTag{
			PostID: postToUpsert.ID,
			Tag:    tag,
		})
	}

	if len(tagsToUpsert) > 0 {
		if err = upsertPostTags(ctx, tx, tagsToUpsert); err != nil {
			return fmt.Errorf("upserting tags: %w", err)
		}
	}

	tagsToDelete := []string{}
	for _, eTag := range existingPost.Tags {
		tagFound := false
		for _, nTag := range request.Tags {
			if nTag == eTag {
				tagFound = true
				break
			}
		}
		if !tagFound {
			tagsToDelete = append(tagsToDelete, eTag)
		}
	}

	if len(tagsToDelete) > 0 {
		if err = deletePostTags(ctx, tx, request.ID, tagsToDelete); err != nil {
			return fmt.Errorf("deleting tags from post: %w", err)
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
