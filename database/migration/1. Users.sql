drop table if exists users;

create table users (
    id uuid not null,
    ext_id varchar(255) unique not null,
    email varchar(128) not null,
    full_name varchar(255) not null,
    profilePicture varchar(255),
    subscribed boolean not null default false,
    primary key (id)
);
