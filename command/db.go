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

  songStmt, songErr := db.Prepare(`insert into songs(album, title, trackNum, discNum, duration,
    flags, full_path, base_path, mime, extension, encoded_extension, is_encoded, md5)
    values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) returning id`)
  checkError(songErr)
  defer songStmt.Close()

  for _, artist := range(m) {
    var artistId int
    err := artistStmt.QueryRow(artist.Name, artist.SortName).Scan(&artistId)
    checkError(err)
    artist.Serial = artistId
    for _, album := range(artist.Albums) {
      var albumId int
      err := albumStmt.QueryRow(artistId, album.Title, album.SortTitle).Scan(&albumId)
      checkError(err)
      album.Serial = albumId
      for _, song := range(album.Songs) {
        var songId int
        err := songStmt.QueryRow(albumId, song.Title, song.TrackNumber, song.DiscNumber, song.Duration,
          song.Flags, song.FullPath, song.BasePath, song.Mime, song.Extension, song.EncodedExtension,
          song.IsEncoded, song.Md5).Scan(&songId)
        checkError(err)
        song.Serial = songId
      }
    }
  }
}

func readArtistMapFromDb() map[string]common.Artist {
  db, err := sql.Open("pgx", config.DbUrl)
  checkError(err)
  defer db.Close()

  artistStmt, artistErr := db.Prepare("select id, name, sortName from artists")
  checkError(artistErr)
  defer artistStmt.Close()

  artistRows, artistErr := artistStmt.Query()
  checkError(artistErr)

  artistMap := make(map[string]common.Artist, 1000)
  for artistRows.Next() {
    var artistId int
    var artistName, artistSortName string
    err := artistRows.Scan(&artistId, &artistName, &artistSortName)
    checkError(err)
    var artist common.Artist
    artist.Serial = artistId
    artist.Name = artistName
    artist.SortName = artistSortName
    artistMap[artistName] = artist
  }
  return artistMap
}
