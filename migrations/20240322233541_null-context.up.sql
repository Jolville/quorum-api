begin;

alter table post alter column context set not null;

commit;