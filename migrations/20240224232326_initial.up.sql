begin;

create table unverified_customer (
    id uuid primary key,
    email text not null,
    first_name text not null,
    last_name text not null,
    created_at timestamptz not null default now(),
    profession varchar(64) not null
);

create unique index idx_unverified_customer_email on unverified_customer(email);

create table customer (
    id uuid primary key,
    email text not null,
    first_name text not null,
    last_name text not null,
    profession varchar(64) not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz not null default now()
);

create unique index idx_customer_email on customer(email);

commit;