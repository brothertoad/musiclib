package command

import (
  "database/sql"
  "fmt"
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
    fmt.Printf("Attempting to insert artist with name %s and sortName %s\n", v.Name, v.SortName)
    res, err := stmt.Exec(v.Name, v.SortName)
    checkError(err)
    lastId, _ := res.LastInsertId()
    fmt.Printf("Artist ID is %d\n", lastId)
  }
}
