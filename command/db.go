package command

import (
  "database/sql"
  "fmt"
  "strconv"
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
    is_encoded, md5, encoded_source_md5, sublibs from songs where album = $1`)
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

func addSongsToDb(songMaps map[string]common.SongMap) {
  db, err := sql.Open("pgx", config.DbUrl)
  checkError(err)
  defer db.Close()

  artistQueryStmt, artistQueryErr := db.Prepare("select id from artists where name = $1")
  checkError(artistQueryErr)
  defer artistQueryStmt.Close()

  artistInsertStmt, artistInsertErr := db.Prepare("insert into artists(name, sortName) values ($1, $2) returning id")
  checkError(artistInsertErr)
  defer artistInsertStmt.Close()

  albumQueryStmt, albumQueryErr := db.Prepare("select id from albums where artist = $1 and title = $2")
  checkError(albumQueryErr)
  defer albumQueryStmt.Close()

  albumInsertStmt, albumInsertErr := db.Prepare("insert into albums(artist, title, sortTitle) values ($1, $2, $3) returning id")
  checkError(albumInsertErr)
  defer albumInsertStmt.Close()

  songInsertStmt, songInsertErr := db.Prepare(`insert into songs(album, title, trackNum, discNum, duration,
    flags, full_path, base_path, mime, extension, encoded_extension, is_encoded, md5)
    values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) returning id`)
  checkError(songInsertErr)
  defer songInsertStmt.Close()

  // For each song, we need to check to see if the artist and album already
  // exist.  If not, we need to add them.
  for _, songMap := range(songMaps) {
    var artistId int
    err := artistQueryStmt.QueryRow(songMap[common.ArtistKey]).Scan(&artistId)
    if err != nil && err != sql.ErrNoRows {
      checkError(err)
    }
    if err != nil {
      // err must be ErrNoRows, so the artist needs to be added.
      fmt.Printf("Need to add artist %s to database\n", songMap[common.ArtistKey])
      err := artistInsertStmt.QueryRow(songMap[common.ArtistKey], songMap[common.ArtistSortKey]).Scan(&artistId)
      checkError(err)
      fmt.Printf("Added artist %s, id is %d\n", songMap[common.ArtistKey], artistId)
    }
    var albumId int
    err = albumQueryStmt.QueryRow(artistId, songMap[common.AlbumKey]).Scan(&albumId)
    if err != nil && err != sql.ErrNoRows {
      checkError(err)
    }
    if err != nil {
      // err must be ErrNoRows, so the album needs to be added.
      fmt.Printf("Need to add album %s by artist %d to database\n", songMap[common.AlbumKey], artistId)
      err := albumInsertStmt.QueryRow(artistId, songMap[common.AlbumKey], songMap[common.AlbumSortKey]).Scan(&albumId)
      checkError(err)
      fmt.Printf("Added album %s, id is %d\n", songMap[common.AlbumKey], albumId)
    }
    // Now we can add the song.
    var songId int
    trackNumber, _ := strconv.Atoi(songMap[common.TrackNumberKey])
    discNumber, _ := strconv.Atoi(songMap[common.DiscNumberKey])
    isEncoded, _ := strconv.ParseBool(songMap[common.IsEncodedKey])
    err = songInsertStmt.QueryRow(albumId, songMap[common.TitleKey], trackNumber, discNumber,
      songMap[common.DurationKey], songMap[common.FlagsKey], songMap[common.FullPathKey],
      songMap[common.BasePathKey], songMap[common.MimeKey], songMap[common.ExtensionKey],
      songMap[common.EncodedExtensionKey], isEncoded, songMap[common.Md5Key]).Scan(&songId)
    checkError(err)
    fmt.Printf("Added song %s, id is %d\n", songMap[common.TitleKey], songId)
  }
}

func deleteSongsFromDb(songMaps map[string]common.SongMap) {
  db, err := sql.Open("pgx", config.DbUrl)
  checkError(err)
  defer db.Close()

  deleteStmt, deleteErr := db.Prepare("delete from songs where id = $1")
  checkError(deleteErr)
  defer deleteStmt.Close()

  for _, songMap := range(songMaps) {
    // Have to convert the serial value in the SongMap from a string to an int.
    serial, _ := strconv.Atoi(songMap[common.SerialKey])
    _, err := deleteStmt.Exec(serial)
    checkError(err)
  }
}

// Delete any albums/artists that don't have any songs.
func deleteEmptyParents() {
  db, err := sql.Open("pgx", config.DbUrl)
  checkError(err)
  defer db.Close()

  albumQueryStmt, albumQueryErr := db.Prepare("select id from albums")
  checkError(albumQueryErr)
  defer albumQueryStmt.Close()

  songQueryStmt, songQueryErr := db.Prepare("select count(*) from songs where album = $1")
  checkError(songQueryErr)
  defer songQueryStmt.Close()

  albumRows, err := albumQueryStmt.Query()
  checkError(err)
  albumsToDelete := make([]int, 0)
  for albumRows.Next() {
    var albumId int
    err := albumRows.Scan(&albumId)
    checkError(err)
    var songCount int
    err = songQueryStmt.QueryRow(albumId).Scan(&songCount)
    checkError(err)
    if songCount == 0 {
      fmt.Printf("Need to delete album %d\n", albumId)
      albumsToDelete = append(albumsToDelete, albumId)
    }
  }
}
