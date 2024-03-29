#!/bin/bash

MUSICDB=musiclib
if [ $# -gt 0 ]; then
    MUSICDB=$1
fi

psql --user postgres --set=MUSICDB=$MUSICDB -f delete_db.sql
psql --user postgres --set=MUSICDB=$MUSICDB -f create_db.sql
psql --user musiclib -f create_tables.sql -d $MUSICDB
