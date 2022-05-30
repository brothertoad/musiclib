drop database if exists musiclib;
drop role if exists musiclib;

create role musiclib login password 'R0ndoAllaTurca';
create database musiclib with owner = musiclib;
