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
	IDs database.UUIDSlice
}

type post struct {
	ID          uuid.UUID          `db:"id"`
	DesignPhase *DesignPhase       `db:"design_phase"`
	Context     *string            `db:"context"`
	Category    *PostCategory      `db:"category"`
	OpensAt     *time.Time         `db:"opens_at"`
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
			post.opens_at,
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
			post.opens_at,
			post.closes_at
		order by post.opens_at desc
	%s`, query, dbLock)

	if err := db.SelectContext(ctx, &posts, query, args...); err != nil {
		return nil, fmt.Errorf("selecting posts: %w", err)
	}

	return posts, nil
}

type getPostOptionsByFilterParams struct {
	IDs     database.UUIDSlice
	PostIDs database.UUIDSlice
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

type getPostVotesByFilterParams struct {
	IDs database.UUIDSlice
}

type postVote struct {
	ID           uuid.UUID `db:"id"`
	PostID       uuid.UUID `db:"post_id"`
	CustomerID   uuid.UUID `db:"voter_id"`
	PostOptionID uuid.UUID `db:"post_option_id"`
	Reason       *string   `db:"reason"`
}

func getPostVotesByFilter(
	ctx context.Context,
	db database.Q,
	params getPostVotesByFilterParams,
	dbLock DBLock,
) ([]postVote, error) {
	postVote := []postVote{}
	query := `
		select
			id,
			post_id,
			voter_id,
			post_option_id,
			reason
		from post_vote
		where true
	`

	args := []any{}
	if len(params.IDs) > 0 {
		args = append(args, params.IDs)
		query = fmt.Sprintf("%s and id = any($%v)", query, len(args))
	}

	query = fmt.Sprintf("%s %s", query, dbLock)

	if err := db.SelectContext(ctx, &postVote, query, args...); err != nil {
		return nil, fmt.Errorf("selecting post_options: %w", err)
	}

	return postVote, nil
}

type upsertPostParams struct {
	ID          uuid.UUID     `db:"id"`
	AuthorID    uuid.UUID     `db:"author_id"`
	DesignPhase *DesignPhase  `db:"design_phase"`
	Context     *string       `db:"context"`
	Category    *PostCategory `db:"category"`
	OpensAt     *time.Time    `db:"opens_at"`
	ClosesAt    *time.Time    `db:"closes_at"`
}

func upsertPost(
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
			opens_at,
			closes_at
		) values (
			:id,
			:author_id,
			:design_phase,
			:context,
			:category,
			:opens_at,
			:closes_at
		) on conflict (id) do update set
			updated_at = now(),
			design_phase = excluded.design_phase,
			context = excluded.context,
			category = excluded.category,
			opens_at = excluded.opens_at,
			closes_at = excluded.closes_at

	`, params); err != nil {
		return fmt.Errorf("inserting post: %w", err)
	}
	return nil
}

type postTag struct {
	PostID uuid.UUID `db:"post_id"`
	Tag    string    `db:"tag"`
}

func upsertPostTags(
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
		return fmt.Errorf("inserting post_tag: %w", err)
	}
	return nil
}

func deletePostTags(
	ctx context.Context,
	db database.Q,
	postID uuid.UUID,
	tags []string,
) error {
	if _, err := db.ExecContext(ctx, `
		delete from post_tag where post_id = $1 and tag = any($2)
		`, postID, tags); err != nil {
		return fmt.Errorf("deleting from post_tag: %w", err)
	}
	return nil
}

func deletePostOptions(
	ctx context.Context,
	db database.Q,
	postOptionIDs database.UUIDSlice,
) error {
	if _, err := db.ExecContext(ctx, `
		delete from post_option where id = any($1)
		`, postOptionIDs); err != nil {
		return fmt.Errorf("deleting from post_option: %w", err)
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

type updatePostOption struct {
	ID       uuid.UUID `db:"id"`
	Position int       `db:"position"`
}

func updatePostOptionPositions(
	ctx context.Context,
	db database.Q,
	postOptions []updatePostOption,
) error {
	values := []string{}
	ids := database.UUIDSlice{}
	for _, po := range postOptions {
		values = append(values, fmt.Sprintf("(%v, %v)", po.ID, po.Position))
		ids = append(ids, po.ID)
	}
	query := fmt.Sprintf(`
		with post_option_update(post_option_id, position) as (values(%s))
		update post_option
			set position = (
				select post_option_update.position
				from post_option_update
				where post_option_update.id = post_option.id
			)
		where id = any($1)`, values,
	)
	if _, err := db.ExecContext(ctx, query, ids); err != nil {
		return fmt.Errorf("updating post_option: %w", err)
	}
	return nil
}
