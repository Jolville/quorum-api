begin;

create type user_profression as enum (
    'PRODUCT_DESIGNER',
    'SOFTWARE_ENGINEER',
    'OTHER'
);

create table unverified_user (
    id uuid primary key,
    email text not null,
    first_name text not null,
    last_name text not null,
    created_at timestamptz not null default now(),
    user_profression user_profression not null
);

create unique index idx_unverified_user_email on unverified_user(email);

create table "user" (
    id uuid primary key,
    email text not null,
    first_name text not null,
    last_name text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz not null default now()
);

commit;