begin;

create table post_tag (
    post_id uuid not null references post(id),
    tag text not null,
    primary key (post_id, tag)
);

create index idx_post_tag_post_id on post_tag(post_id);

commit;