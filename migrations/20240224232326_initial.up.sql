begin;

create table unverified_customer (
    id uuid primary key,
    email text not null,
    first_name text,
    last_name text,
    created_at timestamptz not null default now(),
    profession text
);

create unique index idx_unverified_customer_email on unverified_customer(email);

create table customer (
    id uuid primary key,
    email text not null,
    first_name text,
    last_name text,
    profession text,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create unique index idx_customer_email on customer(email);

commit;