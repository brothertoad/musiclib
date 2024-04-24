package main

import (
  "database/sql"
)

func loadArtistsByState(db *sql.DB, state int) ([]ArtistModel, error) {
  resp := make([]ArtistModel, 0)
  var stmt *sql.Stmt
  var err error
  if state != 0 {
    stmt, err = db.Prepare("select id, name from artists where exists " +
      "(select * from albums where albums.artist = artists.id and exists " +
        "(select * from songs where songs.album = albums.id and state = $1)) order by sort_name")
  } else {
    stmt, err = db.Prepare("select id, name from artists order by sort_name")
  }
  if err != nil {
    return resp, err
  }
  defer stmt.Close()
  var rows *sql.Rows
  if state != 0 {
    rows, err = stmt.Query(state)
  } else {
    rows, err = stmt.Query()
  }
  if err != nil {
    return resp, err
  }
  for rows.Next() {
    var artist ArtistModel
    err := rows.Scan(&artist.Id, &artist.Name)
    if err != nil {
      return resp, err
    }
    resp = append(resp, artist)
  }
  return resp, nil
}
