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

  artistStmt, artistErr := db.Prepare("insert into artists(name, sortName) values ($1, $2) returning id")
  checkError(artistErr)
  defer artistStmt.Close()

  albumStmt, albumErr := db.Prepare("insert into albums(artist, title, sortTitle) values ($1, $2, $3) returning id")
  checkError(albumErr)
  defer albumStmt.Close()

  songStmt, songErr := db.Prepare("insert into songs(title, sortTitle) values ($1, $2) returning id")
  checkError(songErr)
  defer songStmt.Close()

  for _, artist := range(m) {
    var artistId int
    err := artistStmt.QueryRow(artist.Name, artist.SortName).Scan(&artistId)
    checkError(err)
    for _, album := range(artist.Albums) {
      var albumId int
      err := albumStmt.QueryRow(artistId, album.Title, album.SortTitle).Scan(&albumId)
      checkError(err)
    }
  }
}
