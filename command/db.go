package command

import (
  "database/sql"
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
    artist.Id = artistId
    for _, album := range(artist.Albums) {
      var albumId int
      err := albumStmt.QueryRow(artistId, album.Title, album.SortTitle).Scan(&albumId)
      checkError(err)
      album.Id = albumId
      for _, song := range(album.Songs) {
        var songId int
        err := songStmt.QueryRow(albumId, song.Title, song.TrackNumber, song.DiscNumber, song.Duration,
          song.Flags, song.FullPath, song.BasePath, song.Mime, song.Extension, song.EncodedExtension,
          song.IsEncoded, song.Md5).Scan(&songId)
        checkError(err)
        song.Id = songId
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
    err := artistRows.Scan(&artist.Id, &artist.Name, &artist.SortName)
    checkError(err)
    artist.Albums = make(map[string]*common.Album)
    artistMap[artist.Name] = artist
    totalArtists++

    albumRows, albumErr := albumStmt.Query(artist.Id)
    checkError(albumErr)
    for albumRows.Next() {
      album := new(common.Album)
      err := albumRows.Scan(&album.Id, &album.Title, &album.SortTitle)
      checkError(err)
      album.Songs = make([]*common.Song, 0, 100)
      artist.Albums[album.Title] = album
      totalAlbums++

      songRows, songErr := songStmt.Query(album.Id)
      checkError(songErr)
      for songRows.Next() {
        song := new(common.Song)
        err := songRows.Scan(&song.Id, &song.Title, &song.TrackNumber,
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
      err := artistInsertStmt.QueryRow(songMap[common.ArtistKey], songMap[common.ArtistSortKey]).Scan(&artistId)
      checkError(err)
    }
    var albumId int
    err = albumQueryStmt.QueryRow(artistId, songMap[common.AlbumKey]).Scan(&albumId)
    if err != nil && err != sql.ErrNoRows {
      checkError(err)
    }
    if err != nil {
      // err must be ErrNoRows, so the album needs to be added.
      err := albumInsertStmt.QueryRow(artistId, songMap[common.AlbumKey], songMap[common.AlbumSortKey]).Scan(&albumId)
      checkError(err)
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
    // Have to convert the id value in the SongMap from a string to an int.
    id, _ := strconv.Atoi(songMap[common.IdKey])
    _, err := deleteStmt.Exec(id)
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
      albumsToDelete = append(albumsToDelete, albumId)
    }
  }
  deleteIdsFromTable(db, albumsToDelete, "albums")
  deleteEmptyArtists(db)
}

func deleteEmptyArtists(db *sql.DB) {
  artistQueryStmt, err := db.Prepare("select id from artists")
  checkError(err)
  defer artistQueryStmt.Close()
  albumQueryStmt, err := db.Prepare("select count(*) from albums where artist = $1")
  checkError(err)
  defer albumQueryStmt.Close();

  artistRows, err := artistQueryStmt.Query()
  checkError(err)
  artistsToDelete := make([]int, 0)
  for artistRows.Next() {
    var artistId int
    err = artistRows.Scan(&artistId)
    checkError(err)
    var albumCount int
    err = albumQueryStmt.QueryRow(artistId).Scan(&albumCount)
    checkError(err)
    if albumCount == 0 {
      artistsToDelete = append(artistsToDelete, artistId)
    }
  }
  deleteIdsFromTable(db, artistsToDelete, "artists")
}

func deleteIdsFromTable(db *sql.DB, ids []int, table string) {
  if len(ids) == 0 {
    return
  }
  stmt, err := db.Prepare("delete from " + table + " where id = $1")
  checkError(err)
  defer stmt.Close()
  for _, id := range(ids) {
    _, err := stmt.Exec(id)
    checkError(err)
  }
}
