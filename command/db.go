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

  // Insert the artists.
  stmt, err := db.Prepare("insert into artists(name, sortName) values ($1, $2) returning id")
  defer stmt.Close()
  for _, v := range(m) {
    var artistId int
    err := stmt.QueryRow(v.Name, v.SortName).Scan(&artistId)
    checkError(err)
  }
}
