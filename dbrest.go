package main

import (
  "database/sql"
)

func loadArtists(db *sql.DB, state int) ([]ArtistModel, error) {
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

func loadAlbums(db *sql.DB, artistId, state int) ([]AlbumModel, error) {
  resp := make([]AlbumModel, 0)
  var stmt *sql.Stmt
  var err error
  if state != 0 {
    stmt, err = db.Prepare("select id, title from albums where artist = $1 and exists " +
        "(select * from songs where songs.album = albums.id and state = $2) order by sort_title")
  } else {
    stmt, err = db.Prepare("select id, title from albums where artist = $1 order by sort_title")
  }
  if err != nil {
    return resp, err
  }
  defer stmt.Close()
  var rows *sql.Rows
  if state != 0 {
    rows, err = stmt.Query(artistId, state)
  } else {
    rows, err = stmt.Query(artistId)
  }
  if err != nil {
    return resp, err
  }
  for rows.Next() {
    var album AlbumModel
    err := rows.Scan(&album.Id, &album.Title)
    if err != nil {
      return resp, err
    }
    resp = append(resp, album)
  }
  return resp, nil
}

func loadSongs(db *sql.DB, albumId, state int) ([]SongModel, error) {
  resp := make([]SongModel, 0)
  var stmt *sql.Stmt
  var err error
  if state != 0 {
    stmt, err = db.Prepare("select id, track_number, title from songs where album = $1 and state = $2 order by track_number")
  } else {
    stmt, err = db.Prepare("select id, track_number, title from songs where album = $1 order by track_number")
  }
  if err != nil {
    return resp, err
  }
  defer stmt.Close()
  var rows *sql.Rows
  if state != 0 {
    rows, err = stmt.Query(albumId, state)
  } else {
    rows, err = stmt.Query(albumId)
  }
  if err != nil {
    return resp, err
  }
  for rows.Next() {
    var song SongModel
    err := rows.Scan(&song.Id, &song.TrackNum, &song.Title)
    if err != nil {
      return resp, err
    }
    resp = append(resp, song)
  }
  return resp, nil
}

func loadAllSongs(db *sql.DB, state int) ([]SongModel, error) {
  resp := make([]SongModel, 0)
  var stmt *sql.Stmt
  var err error
  if state != 0 {
    stmt, err = db.Prepare("select song.id, song.track_number, song.title, album.title, artist.name from songs song, albums album, artists artist where song.state = $1" +
      " and song.album = album.id and album.artist = artist.id order by artist.sort_name, album.sort_title, song.track_number")
  } else {
    stmt, err = db.Prepare("select song.id, song.track_number, song.title, album.title, artist.name from songs song, albums album, artists artist where" +
      " song.album = album.id and album.artist = artist.id order by artist.sort_name, album.sort_title, song.track_number")
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
    var song SongModel
    err := rows.Scan(&song.Id, &song.TrackNum, &song.Title, &song.Album, &song.Artist)
    if err != nil {
      return resp, err
    }
    resp = append(resp, song)
  }
  return resp, nil
}

func loadSongStates(db *sql.DB, req *UpdateSongStatesModel) error {
  for _, songId := range(req.SongIds) {
    _, err := db.Exec("update songs set state = $1 where id = $2", req.State, songId)
    if err != nil {
      return err
    }
  }
  return nil
}
