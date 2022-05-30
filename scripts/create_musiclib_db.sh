#!/bin/bash

psql --user postgres -f create_db.sql
psql --user musiclib -f create_tables.sql
