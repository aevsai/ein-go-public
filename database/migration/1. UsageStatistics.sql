drop table if exists usage_statistics;

create table usage_statistics (
    id uuid not null,
    user_id uuid not null references users (id),
    date date not null,
    tokens_in int default 0,
    tokens_out int not null default 0,
    primary key (id)
);