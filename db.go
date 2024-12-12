package main

import (
  "database/sql"
  "log"
  "strconv"
  _ "github.com/jackc/pgx/v4/stdlib"
  "github.com/brothertoad/btu"
  "github.com/brothertoad/tags"
)

func getDbConnection() *sql.DB {
  db, err := sql.Open("pgx", config.DbUrl)
  btu.CheckError(err)
  return db
}

func addArtistMapToDb(db *sql.DB, m map[string]Artist) {
  artistStmt, artistErr := db.Prepare("insert into artists(name, sort_name) values ($1, $2) returning id")
  btu.CheckError(artistErr)
  defer artistStmt.Close()

  albumStmt, albumErr := db.Prepare("insert into albums(artist, title, sort_title) values ($1, $2, $3) returning id")
  btu.CheckError(albumErr)
  defer albumStmt.Close()

  songStmt, songErr := db.Prepare(`insert into songs(album, title, track_number, disc_number, duration,
    flags, relative_path, base_path, mime, extension, encoded_extension, is_encoded, md5, size_and_time)
    values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) returning id`)
  btu.CheckError(songErr)
  defer songStmt.Close()

  for _, artist := range(m) {
    var artistId int
    err := artistStmt.QueryRow(artist.Name, artist.SortName).Scan(&artistId)
    if err != nil {
      log.Fatalf("addArtistMapToDb: Error inserting artist '%s', error is %s\n", artist.Name, err.Error())
    }
    artist.Id = artistId
    for _, album := range(artist.Albums) {
      var albumId int
      err := albumStmt.QueryRow(artistId, album.Title, album.SortTitle).Scan(&albumId)
      btu.CheckError(err)
      album.Id = albumId
      for _, song := range(album.Songs) {
        var songId int
        err := songStmt.QueryRow(albumId, song.Title, song.TrackNumber, song.DiscNumber, song.Duration,
          song.Flags, song.RelativePath, song.BasePath, song.Mime, song.Extension, song.EncodedExtension,
          song.IsEncoded, song.Md5, song.SizeAndTime).Scan(&songId)
        if err != nil {
          log.Fatalf("addArtistMapToDb: Error inserting song '%s', album '%s', artist '%s', error is %s\n", song.Title, album.Title, artist.Name, err.Error())
        }
        song.Id = songId
      }
    }
  }
}

func readArtistMapFromDb(db *sql.DB) map[string]Artist {
  artistStmt, artistErr := db.Prepare("select id, name, sort_name from artists")
  btu.CheckError(artistErr)
  defer artistStmt.Close()

  albumStmt, albumErr := db.Prepare("select id, title, sort_title from albums where artist = $1")
  btu.CheckError(albumErr)
  defer albumStmt.Close()

  songStmt, songErr := db.Prepare(`select id, title, track_number, disc_number, duration,
    flags, state, relative_path, base_path, mime, extension, encoded_extension,
    is_encoded, md5, size_and_time, encoded_source, sublibs from songs where album = $1`)
  btu.CheckError(songErr)
  defer songStmt.Close()

  artistRows, artistQueryErr := artistStmt.Query()
  btu.CheckError(artistQueryErr)

  totalArtists := 0
  totalAlbums := 0
  totalSongs := 0

  artistMap := make(map[string]Artist, 1000)
  for artistRows.Next() {
    var artist Artist
    err := artistRows.Scan(&artist.Id, &artist.Name, &artist.SortName)
    btu.CheckError(err)
    artist.Albums = make(map[string]*Album)
    artistMap[artist.Name] = artist
    totalArtists++

    albumRows, albumErr := albumStmt.Query(artist.Id)
    btu.CheckError(albumErr)
    for albumRows.Next() {
      album := new(Album)
      err := albumRows.Scan(&album.Id, &album.Title, &album.SortTitle)
      btu.CheckError(err)
      album.Songs = make([]*Song, 0, 100)
      artist.Albums[album.Title] = album
      totalAlbums++

      songRows, songErr := songStmt.Query(album.Id)
      btu.CheckError(songErr)
      for songRows.Next() {
        song := new(Song)
        err := songRows.Scan(&song.Id, &song.Title, &song.TrackNumber,
          &song.DiscNumber, &song.Duration, &song.Flags, &song.State, &song.RelativePath, &song.BasePath,
          &song.Mime, &song.Extension, &song.EncodedExtension, &song.IsEncoded,
          &song.Md5, &song.SizeAndTime, &song.EncodedSource, &song.Sublibs)
        btu.CheckError(err)
        album.Songs = append(album.Songs, song)
        totalSongs++
      }
    }
  }
  return artistMap
}

func readSongListFromDb(db *sql.DB) []Song {
  songs := make([]Song, 0, 5000)
  stmt, err := db.Prepare(`select id, title, track_number, disc_number, duration,
    flags, state, relative_path, base_path, mime, extension, encoded_extension,
    is_encoded, md5, size_and_time, encoded_source, sublibs from songs`)
  btu.CheckError(err)
  defer stmt.Close()
  rows, err := stmt.Query()
  btu.CheckError(err)
  for rows.Next() {
    var song Song
    err := rows.Scan(&song.Id, &song.Title, &song.TrackNumber,
      &song.DiscNumber, &song.Duration, &song.Flags, &song.State, &song.RelativePath, &song.BasePath,
      &song.Mime, &song.Extension, &song.EncodedExtension, &song.IsEncoded,
      &song.Md5, &song.SizeAndTime, &song.EncodedSource, &song.Sublibs)
    btu.CheckError(err)
    songs = append(songs, song)
  }
  return songs
}

func addSongsToDb(db *sql.DB, songMaps map[string]tags.TagMap) {
  artistQueryStmt, artistQueryErr := db.Prepare("select id from artists where name = $1")
  btu.CheckError(artistQueryErr)
  defer artistQueryStmt.Close()

  artistInsertStmt, artistInsertErr := db.Prepare("insert into artists(name, sort_name) values ($1, $2) returning id")
  btu.CheckError(artistInsertErr)
  defer artistInsertStmt.Close()

  albumQueryStmt, albumQueryErr := db.Prepare("select id from albums where artist = $1 and title = $2")
  btu.CheckError(albumQueryErr)
  defer albumQueryStmt.Close()

  albumInsertStmt, albumInsertErr := db.Prepare("insert into albums(artist, title, sort_title) values ($1, $2, $3) returning id")
  btu.CheckError(albumInsertErr)
  defer albumInsertStmt.Close()

  songInsertStmt, songInsertErr := db.Prepare(`insert into songs(album, title, track_number, disc_number, duration,
    flags, relative_path, base_path, mime, extension, encoded_extension, is_encoded, md5, size_and_time)
    values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) returning id`)
  btu.CheckError(songInsertErr)
  defer songInsertStmt.Close()

  // For each song, we need to check to see if the artist and album already
  // exist.  If not, we need to add them.
  for _, songMap := range(songMaps) {
    var artistId int
    err := artistQueryStmt.QueryRow(songMap[tags.ArtistKey]).Scan(&artistId)
    if err != nil && err != sql.ErrNoRows {
      btu.CheckError(err)
    }
    if err != nil {
      // err must be ErrNoRows, so the artist needs to be added.
      err := artistInsertStmt.QueryRow(songMap[tags.ArtistKey], songMap[tags.ArtistSortKey]).Scan(&artistId)
      btu.CheckError(err)
    }
    var albumId int
    err = albumQueryStmt.QueryRow(artistId, songMap[tags.AlbumKey]).Scan(&albumId)
    if err != nil && err != sql.ErrNoRows {
      btu.CheckError(err)
    }
    if err != nil {
      // err must be ErrNoRows, so the album needs to be added.
      err := albumInsertStmt.QueryRow(artistId, songMap[tags.AlbumKey], songMap[tags.AlbumSortKey]).Scan(&albumId)
      btu.CheckError(err)
    }
    // Now we can add the song.
    var songId int
    trackNumber := btu.Atoi(songMap[tags.TrackNumberKey])
    discNumber := btu.Atoi(songMap[tags.DiscNumberKey])
    isEncoded, _ := strconv.ParseBool(songMap[tags.IsEncodedKey])
    err = songInsertStmt.QueryRow(albumId, songMap[tags.TitleKey], trackNumber, discNumber,
      songMap[tags.DurationKey], songMap[tags.FlagsKey], songMap[tags.RelativePathKey],
      songMap[tags.BasePathKey], songMap[tags.MimeKey], songMap[tags.ExtensionKey],
      songMap[tags.EncodedExtensionKey], isEncoded, songMap[tags.Md5Key], songMap[tags.SizeAndTimeKey]).Scan(&songId)
    btu.CheckError(err)
  }
}

func updateSongEncodedSource(db *sql.DB, song Song) {
  _, err := db.Exec("update songs set encoded_source = $1 where id = $2", song.EncodedSource, song.Id)
  btu.CheckError(err)
}

func updateSongPaths(db *sql.DB, id int, songMap tags.TagMap) {
  _, err := db.Exec("update songs set relative_path = $1, base_path = $2 where id = $3", songMap[tags.RelativePathKey], songMap[tags.BasePathKey], id)
  btu.CheckError(err)
}

func deleteSongsFromDb(db *sql.DB, songMaps map[string]tags.TagMap) {
  deleteStmt, deleteErr := db.Prepare("delete from songs where id = $1")
  btu.CheckError(deleteErr)
  defer deleteStmt.Close()

  for _, songMap := range(songMaps) {
    // Have to convert the id value in the SongMap from a string to an int.
    id := btu.Atoi(songMap[tags.IdKey])
    _, err := deleteStmt.Exec(id)
    btu.CheckError(err)
  }
}

// Delete any albums that don't have any songs and artists that don't have any albums.
func deleteEmptyContainers(db *sql.DB) {
  deleteEmptyParents(db, "albums", "songs", "album")
  deleteEmptyParents(db, "artists", "albums", "artist")
}

func deleteEmptyParents(db *sql.DB, parentTable, childTable, keyCol string) {
  parentQueryStmt, err := db.Prepare("select id from " + parentTable)
  btu.CheckError(err)
  defer parentQueryStmt.Close()
  childQueryStmt, err := db.Prepare("select count(*) from " + childTable + " where " + keyCol + " = $1")
  btu.CheckError(err)
  defer childQueryStmt.Close();

  rows, err := parentQueryStmt.Query()
  btu.CheckError(err)
  idsToDelete := make([]int, 0)
  for rows.Next() {
    var id int
    err = rows.Scan(&id)
    btu.CheckError(err)
    var count int
    err = childQueryStmt.QueryRow(id).Scan(&count)
    btu.CheckError(err)
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
  btu.CheckError(err)
  defer stmt.Close()
  for _, id := range(ids) {
    _, err := stmt.Exec(id)
    btu.CheckError(err)
  }
}
