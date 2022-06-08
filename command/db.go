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

  albumStmt, albumErr := db.Prepare("select id, title, sortTitle from albums where artist = $1")
  checkError(albumErr)
  defer albumStmt.Close()

  songStmt, songErr := db.Prepare(`select id, title, trackNum, discNum, duration,
    flags, full_path, base_path, mime, extension, encoded_extension,
    is_encoded, md5, encoded_md5, sublibs from songs where album = $1`)
  checkError(songErr)
  defer songStmt.Close()

  artistRows, artistQueryErr := artistStmt.Query()
  checkError(artistQueryErr)

  totalArtists := 0
  totalAlbums := 0
  totalSongs := 0

  artistMap := make(map[string]common.Artist, 1000)
  for artistRows.Next() {
    var artist common.Artist
    err := artistRows.Scan(&artist.Serial, &artist.Name, &artist.SortName)
    checkError(err)
    artist.Albums = make(map[string]*common.Album)
    artistMap[artist.Name] = artist
    totalArtists++

    albumRows, albumErr := albumStmt.Query(artist.Serial)
    checkError(albumErr)
    for albumRows.Next() {
      album := new(common.Album)
      err := albumRows.Scan(&album.Serial, &album.Title, &album.SortTitle)
      checkError(err)
      album.Songs = make([]*common.Song, 0, 100)
      artist.Albums[album.Title] = album
      totalAlbums++

      songRows, songErr := songStmt.Query(album.Serial)
      checkError(songErr)
      for songRows.Next() {
        song := new(common.Song)
        err := songRows.Scan(&song.Serial, &song.Title, &song.TrackNumber,
          &song.DiscNumber, &song.Duration, &song.Flags, &song.FullPath, &song.BasePath,
          &song.Mime, &song.Extension, &song.EncodedExtension, &song.IsEncoded,
          &song.Md5, &song.EncodedSourceMd5, &song.Sublibs)
        checkError(err)
        album.Songs = append(album.Songs, song)
        totalSongs++
      }
    }
  }
  return artistMap
}
