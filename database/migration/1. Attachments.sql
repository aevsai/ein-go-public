drop table if exists attachments;

create table attachments (
    id uuid not null,
    message_id uuid references messages (id),
    bucket varchar not null default 'userfiles',
    key varchar not null,
    created_at timestamp not null default now(),
    primary key (id) 
);
