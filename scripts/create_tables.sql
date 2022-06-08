create table artists (
id serial primary key,
name text,
sortName text
);
alter sequence artists_id_seq restart with 100001;

create table albums (
id serial primary key,
artist integer references artists on delete cascade,
title text,
sortTitle text
);
alter sequence albums_id_seq restart with 200001;

create table songs (
id serial primary key,
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
sublibs text default ''
);
alter sequence songs_id_seq restart with 300001;
