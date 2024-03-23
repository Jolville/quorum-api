begin;

alter table post add column live_at timestamptz;

alter table post drop column opens_at;

commit;