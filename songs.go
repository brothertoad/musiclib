package main

import (
  "encoding/hex"
  "fmt"
  "io"
  "io/fs"
  "io/ioutil"
  "log"
  "os"
  "path"
  "path/filepath"
  "sort"
  "strconv"
  "strings"
  "gopkg.in/yaml.v3"
  "github.com/brothertoad/btu"
  "github.com/brothertoad/tags"
)

// In addition to being required, these are the only keys we save in the yaml file.
var requiredKeys = []string {
  tags.TitleKey, tags.ArtistKey, tags.AlbumKey, tags.TrackNumberKey, tags.DiscNumberKey,
  tags.ArtistSortKey, tags.AlbumSortKey, tags.RelativePathKey, tags.BasePathKey,
  tags.MimeKey, tags.ExtensionKey, tags.EncodedExtensionKey, tags.IsEncodedKey,
  tags.FlagsKey, tags.DurationKey, tags.Md5Key, tags.SizeAndTimeKey,
}

////////////////////////////////////////////////////////////////////////
//
// Logic for reading the library.
//
////////////////////////////////////////////////////////////////////////

func loadSongMapSliceFromMusicDir() tags.TagMapSlice {
  songMaps := make(tags.TagMapSlice, 0, 5000)
  filepath.WalkDir(config.MusicDir, func(path string, de fs.DirEntry, err error) error {
    if de.IsDir() {
      return nil
    }
    song := tags.GetStandardTagsFromFile(path)
    if song == nil || len(song) == 0 {
      return nil
    }
    setPaths(song, path)
    song[tags.FlagsKey] = ""
    addSortKeys(song)
    addMd5Key(song)
    info, err := de.Info()
    btu.CheckError2(err, "Couldn't get fileInfo for '%s'\n", path)
    song[tags.SizeAndTimeKey] = fmt.Sprintf("%d-%d", info.Size(), info.ModTime().Unix())
    checkForMissingKeys(song)
    songMaps = append(songMaps, filterKeys(song))
    return nil
  })
  sort.Sort(songMaps)
  return songMaps
}

func setPaths(song tags.TagMap, path string) {
  relativePath := path[musicDirLength:]
  song[tags.RelativePathKey] = relativePath
  // Remove the extension to get the base path.
  pathLength := len(relativePath)
  extLength := len(song[tags.ExtensionKey])
  song[tags.BasePathKey] = relativePath[0:(pathLength-extLength)]
}

func addSortKeys(song tags.TagMap) {
  addSortKey(song, tags.ArtistKey, tags.ArtistSortKey)
  addSortKey(song, tags.AlbumKey, tags.AlbumSortKey)
}

func addSortKey(song tags.TagMap, pureKey string, sortKey string) {
  if _, present := song[sortKey]; !present {
    if vp, purePresent := song[pureKey]; purePresent {
      song[sortKey] = getSortValue(vp)
    } else {
      // If we don't have the pure key, we've got problems.
      log.Fatalf("no key '%s' for '%s'\n", pureKey, song[tags.RelativePathKey])
    }
  }
}

func getSortValue(pureValue string) string {
  if strings.HasPrefix(pureValue, "A ") {
    return pureValue[2:]
  }
  if strings.HasPrefix(pureValue, "a ") {
    return pureValue[2:]
  }
  if strings.HasPrefix(pureValue, "An ") {
    return pureValue[3:]
  }
  if strings.HasPrefix(pureValue, "an ") {
    return pureValue[3:]
  }
  if strings.HasPrefix(pureValue, "The ") {
    return pureValue[4:]
  }
  if strings.HasPrefix(pureValue, "the ") {
    return pureValue[4:]
  }
  return pureValue
}

func addMd5Key(song tags.TagMap) {
  f, err := os.Open(path.Join(config.MusicDir,song[tags.RelativePathKey]))
  btu.CheckError(err)
  defer f.Close()
  hasher.Reset()
  if _, err := io.Copy(hasher, f); err != nil {
    log.Fatalf("Error trying to compute md5sum of %s\n", song[tags.RelativePathKey])
  }
  song[tags.Md5Key] = hex.EncodeToString(hasher.Sum(nil))
}

func checkForMissingKeys(song tags.TagMap) {
  for _, k := range(requiredKeys) {
    if _, present := song[k]; !present {
      fmt.Printf("%+v is missing %s\n", song, k)
    }
  }
}

func filterKeys(song tags.TagMap) tags.TagMap {
  filtered := make(tags.TagMap)
  for _, k := range(requiredKeys) {
    filtered[k] = song[k]
  }
  return filtered;
}

////////////////////////////////////////////////////////////////////////
//
// Logic for converting a SongMapSlice to or from the tree-structure
// containing artists, albums and songs.
//
////////////////////////////////////////////////////////////////////////

func songMapsToArtistMap(songMaps tags.TagMapSlice, printStats bool) map[string]Artist {
  artists := make(map[string]Artist)
  // Build a map of artists.
  for _, sm := range(songMaps) {
    name := sm[tags.ArtistKey]
    if _, present := artists[name]; !present {
      var artist Artist
      artist.Name = name
      artist.SortName = sm[tags.ArtistSortKey]
      artist.Albums = make(map[string]*Album)
      artists[name] = artist
    }
  }
  if printStats {
    fmt.Printf("Found %d artists.\n", len(artists))
  }
  // Build the maps of albums.
  numAlbums := 0
  for _, sm := range(songMaps) {
    name := sm[tags.ArtistKey]
    artist := artists[name]
    albumTitle := sm[tags.AlbumKey]
    if _, present := artist.Albums[albumTitle]; !present {
      album := new(Album)
      album.Title = albumTitle
      album.SortTitle = sm[tags.AlbumSortKey]
      album.Songs = make([]*Song, 0, 100)
      artist.Albums[albumTitle] = album
      numAlbums++
    }
  }
  if printStats {
    fmt.Printf("Found %d albums.\n", numAlbums)
  }
  // Build the lists of songs.  Note that we assume the songMap slice is sorted.
  for _, sm := range(songMaps) {
    name := sm[tags.ArtistKey]
    artist := artists[name]
    albumTitle := sm[tags.AlbumKey]
    album := artist.Albums[albumTitle]
    song := new(Song)
    song.Title = sm[tags.TitleKey]
    song.TrackNumber = btu.Atoi2(sm[tags.TrackNumberKey], "Error getting track number for %s\n", song.Title)
    song.DiscNumber = btu.Atoi2(sm[tags.DiscNumberKey], "Error getting disc number for %s\n", song.Title)
    song.Duration = sm[tags.DurationKey]
    song.Mime = sm[tags.MimeKey]
    song.Extension = sm[tags.ExtensionKey]
    song.EncodedExtension = sm[tags.EncodedExtensionKey]
    song.RelativePath = sm[tags.RelativePathKey]
    song.BasePath = sm[tags.BasePathKey]
    song.IsEncoded, _ = strconv.ParseBool(sm[tags.IsEncodedKey])
    song.Flags = sm[tags.FlagsKey]
    song.Md5 = sm[tags.Md5Key]
    song.SizeAndTime = sm[tags.SizeAndTimeKey]
    song.EncodedSource = sm[tags.EncodedSourceKey]
    album.Songs = append(album.Songs, song)
  }
  // Sort the song slice for each album.
  for _, artist := range(artists) {
    for _, album := range(artist.Albums) {
      SortSongSlice(album.Songs)
    }
  }
  return artists
}

func artistMapToSongMaps(artistMap map[string]Artist) tags.TagMapSlice {
  songMaps := make(tags.TagMapSlice, 0, 5000)
  for _, artist := range(artistMap) {
    for _, album := range(artist.Albums) {
      for _, song := range(album.Songs) {
        songMap := make(tags.TagMap, 0)
        songMap[tags.IdKey] = strconv.Itoa(song.Id)
        songMap[tags.TitleKey] = song.Title
        songMap[tags.ArtistKey] = artist.Name
        songMap[tags.AlbumKey] = album.Title
        songMap[tags.TrackNumberKey] = strconv.Itoa(song.TrackNumber)
        songMap[tags.DiscNumberKey] = strconv.Itoa(song.DiscNumber)
        songMap[tags.ArtistSortKey] = artist.SortName
        songMap[tags.AlbumSortKey] = album.SortTitle
        songMap[tags.RelativePathKey] = song.RelativePath
        songMap[tags.BasePathKey] = song.BasePath
        songMap[tags.MimeKey] = song.Mime
        songMap[tags.ExtensionKey] = song.Extension
        songMap[tags.EncodedExtensionKey] = song.EncodedExtension
        songMap[tags.IsEncodedKey] = strconv.FormatBool(song.IsEncoded)
        songMap[tags.FlagsKey] = song.Flags
        songMap[tags.DurationKey] = song.Duration
        songMap[tags.Md5Key] = song.Md5
        songMap[tags.SizeAndTimeKey] = song.SizeAndTime
        songMaps = append(songMaps, songMap)
      }
    }
  }
  sort.Sort(songMaps)
  return songMaps
}

////////////////////////////////////////////////////////////////////////
//
// Logic for reading and writing a SongMapSlice from/to a YAML file.
//
////////////////////////////////////////////////////////////////////////

func loadSongsFromYaml(path string) tags.TagMapSlice {
  songMaps := make(tags.TagMapSlice, 0, 5000)
  // logic came from https://zetcode.com/golang/yaml/
  yfile, err := ioutil.ReadFile(path)
  btu.CheckError(err)
  err2 := yaml.Unmarshal(yfile, &songMaps)
  btu.CheckError(err2)
  return songMaps
}

func saveSongsToYaml(path string, songMaps tags.TagMapSlice) {
  fmt.Printf("Saving yaml in '%s'\n", path)
  data, err := yaml.Marshal(&songMaps)
  btu.CheckError(err)
  err = ioutil.WriteFile(path, data, 0644)
  btu.CheckError(err)
}
