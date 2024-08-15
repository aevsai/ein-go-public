drop table if exists tools;

create table tools (
    id uuid not null,
    name varchar not null,
    host varchar not null,
    method varchar not null,
    params jsonb not null,
    created_at timestamp not null default now(),
    primary key (id)
);