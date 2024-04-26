#!/bin/bash

MUSICDB=musiclib
if [ $# -gt 0 ]; then
    MUSICDB=$1
fi

SCRIPTDIR=`dirname -- "$( readlink -f -- "$0"; )"`

psql --user postgres --set=MUSICDB=$MUSICDB -f $SCRIPTDIR/delete_db.sql
psql --user postgres --set=MUSICDB=$MUSICDB -f $SCRIPTDIR/create_db.sql
psql --user musiclib -f $SCRIPTDIR/create_tables.sql -d $MUSICDB
