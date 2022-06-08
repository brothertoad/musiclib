create table artists (
id int generated always as identity (start with 10001) primary key,
name text,
sortName text
);

create table albums (
id int generated always as identity (start with 20001) primary key,
artist integer references artists on delete cascade,
title text,
sortTitle text,
unique (artist, title)
);

create table songs (
id int generated always as identity (start with 30001) primary key,
album integer references albums on delete cascade,
title text,
sortTitle text,
trackNum integer,
discNum integer,
duration varchar(12),
flags text,
full_path text,
base_path text,
mime varchar(32),
extension varchar(8),
encoded_extension varchar(8),
is_encoded boolean default false,
md5 varchar(40),
encoded_source_md5 varchar(40) default '',
sublibs text default '',
unique (album, trackNum, discNum)
);
