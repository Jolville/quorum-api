begin;

alter table post drop column category;
drop type post_category;

create type post_category as enum (
    'ANIMATION',
    'BRANDING',
    'ILLUSTRATION',
    'PRINT',
    'PRODUCT',
    'TYPOGRAPHY',
    'WEB'
);

alter table post add column category post_category;

commit;