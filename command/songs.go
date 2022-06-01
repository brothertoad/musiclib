package command

import (
  "fmt"
  "strconv"
  "github.com/brothertoad/musiclib/common"
)

func songMapsToArtistMap(songMaps common.SongMapSlice) map[string]common.Artist {
  artists := make(map[string]common.Artist)
  // Build a map of artists.
  for _, sm := range(songMaps) {
    name := sm[common.ArtistKey]
    if _, present := artists[name]; !present {
      var artist common.Artist
      artist.Name = name
      artist.SortName = sm[common.ArtistSortKey]
      artist.Albums = make(map[string]*common.Album)
      artists[name] = artist
    }
  }
  fmt.Printf("Found %d artists.\n", len(artists))
  // Build the maps of albums.
  numAlbums := 0
  for _, sm := range(songMaps) {
    name := sm[common.ArtistKey]
    artist := artists[name]
    albumTitle := sm[common.AlbumKey]
    if _, present := artist.Albums[albumTitle]; !present {
      // var album common.Album
      album := new(common.Album)
      album.Title = albumTitle
      album.SortTitle = sm[common.AlbumSortKey]
      album.Songs = make([]*common.Song, 0, 100)
      artist.Albums[albumTitle] = album
      numAlbums++
    }
  }
  fmt.Printf("Found %d albums.\n", numAlbums)
  // Build the lists of songs.  Note that we assume the songMap slice is sorted.
  for _, sm := range(songMaps) {
    name := sm[common.ArtistKey]
    artist := artists[name]
    albumTitle := sm[common.AlbumKey]
    album := artist.Albums[albumTitle]
    // var song common.Song
    song := new(common.Song)
    song.Title = sm[common.TitleKey]
    song.TrackNumber, _ = strconv.Atoi(sm[common.TrackNumberKey])
    song.DiscNumber, _ = strconv.Atoi(sm[common.DiscNumberKey])
    song.Duration = sm[common.DurationKey]
    song.Mime = sm[common.MimeKey]
    song.Extension = sm[common.ExtensionKey]
    song.EncodedExtension = sm[common.EncodedExtensionKey]
    song.FullPath = sm[common.FullPathKey]
    song.BasePath = sm[common.BasePathKey]
    song.IsEncoded, _ = strconv.ParseBool(sm[common.IsEncodedKey])
    song.Flags = sm[common.FlagsKey]
    song.Md5 = sm[common.Md5Key]
    song.EncodedSourceMd5 = sm[common.EncodedSourceMd5Key]
    album.Songs = append(album.Songs, song)
  }
  return artists
}
