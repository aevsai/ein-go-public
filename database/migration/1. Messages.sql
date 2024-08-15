drop table if exists messages;

create table messages (
    id uuid not null,
    user_id uuid not null references users (id),
    content varchar not null,
    summarized varchar,
    role varchar(255) not null,
    in_context bool default true,
    in_context_by_force bool default false,
    created_at timestamp not null default now(),
    primary key (id)
);