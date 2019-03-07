create table if not exists users (
    id int not null auto_increment primary key,
    email varchar(255) not null,
    pass_hash varbinary(72) not null,
    user_name varchar(255) not null,
    first_name varchar(64) null,
    last_name varchar(128) null,
    photo_url varchar(2083) not null
);

create unique index email_uniq
on users (email);

create unique index usr_name_uniq
on users (user_name);

create table if not exists signin ( 
    id int not null,
    signin_time datetime not null,
    ip varchar(15) not null 
)