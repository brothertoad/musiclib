package command

import (
  "database/sql"
  "strconv"
  _ "github.com/jackc/pgx/v4/stdlib"
  "github.com/brothertoad/musiclib/common"
)

func getDbConnection() *sql.DB {
  db, err := sql.Open("pgx", config.DbUrl)
  checkError(err)
  return db
}

func addArtistMapToDb(db *sql.DB, m map[string]common.Artist) {
  artistStmt, artistErr := db.Prepare("insert into artists(name, sortName) values ($1, $2) returning id")
  checkError(artistErr)
  defer artistStmt.Close()

  albumStmt, albumErr := db.Prepare("insert into albums(artist, title, sortTitle) values ($1, $2, $3) returning id")
  checkError(albumErr)
  defer albumStmt.Close()

  songStmt, songErr := db.Prepare(`insert into songs(album, title, trackNum, discNum, duration,
    flags, relative_path, base_path, mime, extension, encoded_extension, is_encoded, md5)
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
          song.Flags, song.RelativePath, song.BasePath, song.Mime, song.Extension, song.EncodedExtension,
          song.IsEncoded, song.Md5).Scan(&songId)
        checkError(err)
        song.Id = songId
      }
    }
  }
}

func readArtistMapFromDb(db *sql.DB) map[string]common.Artist {
  artistStmt, artistErr := db.Prepare("select id, name, sortName from artists")
  checkError(artistErr)
  defer artistStmt.Close()

  albumStmt, albumErr := db.Prepare("select id, title, sortTitle from albums where artist = $1")
  checkError(albumErr)
  defer albumStmt.Close()

  songStmt, songErr := db.Prepare(`select id, title, trackNum, discNum, duration,
    flags, relative_path, base_path, mime, extension, encoded_extension,
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
          &song.DiscNumber, &song.Duration, &song.Flags, &song.RelativePath, &song.BasePath,
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

func readSongListFromDb(db *sql.DB) []common.Song {
  songs := make([]common.Song, 0, 5000)
  stmt, err := db.Prepare(`select id, title, trackNum, discNum, duration,
    flags, relative_path, base_path, mime, extension, encoded_extension,
    is_encoded, md5, encoded_source_md5, sublibs from songs`)
  checkError(err)
  defer stmt.Close()
  rows, err := stmt.Query()
  checkError(err)
  for rows.Next() {
    var song common.Song
    err := rows.Scan(&song.Id, &song.Title, &song.TrackNumber,
      &song.DiscNumber, &song.Duration, &song.Flags, &song.RelativePath, &song.BasePath,
      &song.Mime, &song.Extension, &song.EncodedExtension, &song.IsEncoded,
      &song.Md5, &song.EncodedSourceMd5, &song.Sublibs)
    checkError(err)
    songs = append(songs, song)
  }
  return songs
}

func addSongsToDb(db *sql.DB, songMaps map[string]common.SongMap) {
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
    flags, relative_path, base_path, mime, extension, encoded_extension, is_encoded, md5)
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
      songMap[common.DurationKey], songMap[common.FlagsKey], songMap[common.RelativePathKey],
      songMap[common.BasePathKey], songMap[common.MimeKey], songMap[common.ExtensionKey],
      songMap[common.EncodedExtensionKey], isEncoded, songMap[common.Md5Key]).Scan(&songId)
    checkError(err)
  }
}

func updateSongEncodedSourceMd5(db *sql.DB, song common.Song) {
  _, err := db.Exec("update songs set encoded_source_md5 = $1 where id = $2", song.EncodedSourceMd5, song.Id)
  checkError(err)
}

func deleteSongsFromDb(db *sql.DB, songMaps map[string]common.SongMap) {
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

// Delete any albums that don't have any songs and artists that don't have any albums.
func deleteEmptyContainers(db *sql.DB) {
  deleteEmptyParents(db, "albums", "songs", "album")
  deleteEmptyParents(db, "artists", "albums", "artist")
}

func deleteEmptyParents(db *sql.DB, parentTable, childTable, keyCol string) {
  parentQueryStmt, err := db.Prepare("select id from " + parentTable)
  checkError(err)
  defer parentQueryStmt.Close()
  childQueryStmt, err := db.Prepare("select count(*) from " + childTable + " where " + keyCol + " = $1")
  checkError(err)
  defer childQueryStmt.Close();

  rows, err := parentQueryStmt.Query()
  checkError(err)
  idsToDelete := make([]int, 0)
  for rows.Next() {
    var id int
    err = rows.Scan(&id)
    checkError(err)
    var count int
    err = childQueryStmt.QueryRow(id).Scan(&count)
    checkError(err)
    if count == 0 {
      idsToDelete = append(idsToDelete, id)
    }
  }
  deleteIdsFromTable(db, idsToDelete, parentTable)
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
