package srvpost

import (
	"context"
	"fmt"
	"quorum-api/database"
	"time"

	"github.com/google/uuid"
)

type DBLock string

const (
	DBLockUnspecified DBLock = ""
	DBLockForUpdate   DBLock = "for update"
)

type PostCategory string

const (
	PostCategoryProductDesign PostCategory = "PRODUCT_DESIGN"
)

type DesignPhase string

const (
	DesignPhaseWireframe DesignPhase = "WIREFRAME"
	DesignPhaseLoFi      DesignPhase = "LO_FI"
	DesignPhaseHiFi      DesignPhase = "HI_FI"
)

type getPostsByFilterParams struct {
	IDs []uuid.UUID
}

type post struct {
	ID          uuid.UUID          `db:"id"`
	DesignPhase *DesignPhase       `db:"design_phase"`
	Context     *string            `db:"context"`
	Category    *PostCategory      `db:"category"`
	LiveAt      *time.Time         `db:"live_at"`
	ClosesAt    *time.Time         `db:"closes_at"`
	AuthorID    uuid.UUID          `db:"author_id"`
	Tags        []string           `db:"tags"`
	OptionIDs   database.UUIDSlice `db:"option_ids"`
	VoteIDs     database.UUIDSlice `db:"vote_ids"`
}

func getPostsByFilter(
	ctx context.Context,
	db database.Q,
	params getPostsByFilterParams,
	dbLock DBLock,
) ([]post, error) {
	posts := []post{}
	query := `
		select
			post.id,
			post.author_id,
			post.context,
			post.category,
			post.created_at,
			post.updated_at,
			post.live_at,
			post.closes_at,
			array_agg(po.id) option_ids,
			array_agg(pv.id) vote_ids,
			array_agg(pt.tag) tags
		from post
		left join post_option po on po.post_id = post.id
		left join post_vote pv on pv.post_id = post.id
		left join post_tag pt on pt.post_id = post.id
		where true
	`

	args := []any{}
	if len(params.IDs) > 0 {
		args = append(args, params.IDs)
		query = fmt.Sprintf("%s and id = any($%v)", query, len(args))
	}

	query = fmt.Sprintf(`%s
		group by
			post.id,
			post.author_id,
			post.context,
			post.category,
			post.created_at,
			post.updated_at,
			post.live_at,
			post.closes_at
		order by post.live_at desc
	%s`, query, dbLock)

	if err := db.SelectContext(ctx, &posts, query, args...); err != nil {
		return nil, fmt.Errorf("selecting posts: %w", err)
	}

	return posts, nil
}

type getPostOptionsByFilterParams struct {
	IDs     []uuid.UUID
	PostIDs []uuid.UUID
}

type postOption struct {
	ID       uuid.UUID `db:"id"`
	PostID   uuid.UUID `db:"post_id"`
	Position int       `db:"position"`
	FileRef  string    `db:"file_ref"`
}

func getPostOptionsByFilter(
	ctx context.Context,
	db database.Q,
	params getPostOptionsByFilterParams,
	dbLock DBLock,
) ([]postOption, error) {
	postOptions := []postOption{}
	query := `
		select
			id,
			post_id,
			position,
			file_ref
		from post_option
		where true
	`

	args := []any{}
	if len(params.IDs) > 0 {
		args = append(args, params.IDs)
		query = fmt.Sprintf("%s and id = any($%v)", query, len(args))
	}
	if len(params.PostIDs) > 0 {
		args = append(args, params.IDs)
		query = fmt.Sprintf("%s and post_id = any($%v)", query, len(args))
	}

	query = fmt.Sprintf(`%s
		order by post_id, position
	%s`, query, dbLock)

	if err := db.SelectContext(ctx, &postOptions, query, args...); err != nil {
		return nil, fmt.Errorf("selecting post_options: %w", err)
	}

	return postOptions, nil
}

type upsertPostParams struct {
	ID          uuid.UUID     `db:"id"`
	AuthorID    uuid.UUID     `db:"author_id"`
	DesignPhase *DesignPhase  `db:"design_phase"`
	Context     *string       `db:"context"`
	Category    *PostCategory `db:"category"`
	LiveAt      *time.Time    `db:"live_at"`
	ClosesAt    *time.Time    `db:"closes_at"`
}

func insertPost(
	ctx context.Context,
	db database.Q,
	params upsertPostParams,
) error {
	if _, err := db.NamedExecContext(ctx, `
		insert into post (
			id,
			author_id,
			design_phase,
			context,
			category,
			live_at,
			closes_at
		) values (
			:id,
			:author_id,
			:design_phase,
			:context,
			:category,
			:live_at,
			:closes_at
		)
	`, params); err != nil {
		return fmt.Errorf("inserting post: %w", err)
	}
	return nil
}

type postTag struct {
	PostID uuid.UUID `db:"post_id"`
	Tag    string    `db:"tag"`
}

func insertPostTags(
	ctx context.Context,
	db database.Q,
	postTags []postTag,
) error {
	if _, err := db.NamedExecContext(ctx, `
		insert into post_tag (
			post_id,
			tag
		) values (
			:post_id,
			:tag
		) on conflict do nothing
	`, postTags); err != nil {
		return fmt.Errorf("inserting post: %w", err)
	}
	return nil
}

func insertPostOptions(
	ctx context.Context,
	db database.Q,
	postOptions []postOption,
) error {
	if _, err := db.NamedExecContext(ctx, `
		insert into post_option (
			id,
			post_id,
			position,
			file_ref
		) values (
			:id,
			:post_id,
			:position,
			:file_ref
		)
	`, postOptions); err != nil {
		return fmt.Errorf("inserting post: %w", err)
	}
	return nil
}
