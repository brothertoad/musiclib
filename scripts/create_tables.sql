create table artists (
id serial primary key,
name varchar(100),
sortName varchar(100)
);
alter sequence artists_id_seq restart with 100001;

create table albums (
id serial primary key,
artist integer references artists on delete cascade,
title varchar(100),
sortTitle varchar(100)
);
alter sequence albums_id_seq restart with 200001;

create table songs (
id serial primary key,
album integer references albums on delete cascade,
title varchar(200),
sortTitle varchar(200),
trackNum integer,
discNum integer,
duration varchar(12),
flags varchar(64),
full_path varchar(300),
base_path varchar(300),
mime varchar(32),
extension varchar(8),
encoded_extension varchar(8),
is_encoded boolean default false,
md5 varchar(40),
encoded_source_md5 varchar(40) default '',
sublibs varchar(256) default ''
);
alter sequence songs_id_seq restart with 300001;
