package main

import (
  "database/sql"
)

func loadArtistsByState(db *sql.DB, state int) ([]ArtistModel, error) {
  resp := make([]ArtistModel, 0)
  stmt, err := db.Prepare("select id, name from artists where exists " +
    "(select * from albums where albums.artist = artists.id and exists " +
      "(select * from songs where songs.album = albums.id and state = $1)) order by sort_name")
  if err != nil {
    return resp, err
  }
  defer stmt.Close()
  rows, err := stmt.Query(state)
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
