begin;

create type design_phase as enum (
    'WIREFRAME',
    'LO_FI',
    'HI_FI'
);

create type post_category as enum (
    'PRODUCT_DESIGN'
);

create table post (
    id uuid primary key,
    author_id uuid not null references customer(id),
    context text not null,
    category post_category,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    live_at timestamptz,
    closes_at timestamptz
);

create index idx_post_author_id on post (author_id);

create table post_tag (
    post_id uuid not null references post(id),
    tag text not null,
    primary key (post_id, tag)
);

create index idx_post_tag_post_id on post_tag(post_id);

create table post_option (
    id uuid primary key,
    post_id uuid not null references post(id),
    file_ref text not null,
    position int not null check (position >= 0),
    unique (post_id, file_ref),
    unique (post_id, position)
);

create index idx_post_option_post_id on post_option(post_id);

create table post_vote (
    id uuid primary key,
    post_id uuid not null references post(id),
    post_option_id uuid not null references post_option(id),
    customer_id uuid not null references customer(id),
    reason text,
    unique (post_id, customer_id)
);

create index idx_post_vote_post_id on post_vote(post_id);
create index idx_post_vote_post_option_id on post_vote(post_option_id);
create index idx_post_vote_customer_id on post_vote(customer_id);

commit;