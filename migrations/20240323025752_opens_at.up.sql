begin;

alter table post drop column live_at;

alter table post add column opens_at timestamptz;

commit;