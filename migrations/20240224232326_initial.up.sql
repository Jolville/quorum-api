begin;

create table unverified_user (
    id uuid primary key,
    email text not null,
    first_name text not null,
    last_name text not null,
    created_at timestamptz not null default now(),
    user_profression varchar(64) not null
);

create unique index idx_unverified_user_email on unverified_user(email);

create table "user" (
    id uuid primary key,
    email text not null,
    first_name text not null,
    last_name text not null,
    user_profression varchar(64) not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz not null default now()
);

commit;