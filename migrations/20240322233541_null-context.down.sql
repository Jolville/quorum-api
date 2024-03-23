begin;

alter table post alter column context drop not null;

commit;