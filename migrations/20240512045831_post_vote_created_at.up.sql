begin;

alter table post_vote add column created_at timestamptz not null default now();

commit;