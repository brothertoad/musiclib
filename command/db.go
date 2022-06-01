package command

import (
  "database/sql"
  _ "github.com/jackc/pgx/v4/stdlib"
  "github.com/brothertoad/musiclib/common"
)

func addArtistMapToDb(m map[string]common.Artist) {
  db, err := sql.Open("pgx", config.DbUrl)
  checkError(err)
  defer db.Close()
}
