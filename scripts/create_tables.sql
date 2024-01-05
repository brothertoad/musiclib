create table artists (
id int generated always as identity (start with 10001) primary key,
name text,
sort_name text,
unique (name)
);

create table albums (
id int generated always as identity (start with 20001) primary key,
artist integer references artists on delete cascade,
title text,
sort_title text,
unique (artist, title)
);

create table songs (
id int generated always as identity (start with 30001) primary key,
album integer references albums on delete cascade,
title text,
sortTitle text,
track_number integer,
disc_number integer,
duration text,
flags text,
relative_path text,
base_path text,
mime text,
extension text,
encoded_extension text,
is_encoded boolean default false,
md5 text,
size bigint,
mod_time bigint,
encoded_source_md5 text default '',
sublibs text default '',
unique (album, track_number, disc_number)
);
