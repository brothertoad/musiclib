package command

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
  "github.com/brothertoad/musiclib/common"
)

var keyTranslations = map[string]string {
  "\xa9nam": tags.TitleKey,
  "\xa9ART" : tags.ArtistKey,
  "\xa9alb" : tags.AlbumKey,
  "soar" : tags.ArtistSortKey,
  "soal" : tags.AlbumSortKey,
  "ALBUM" : tags.AlbumKey,
  "ARTIST" : tags.ArtistKey,
  "TITLE" : tags.TitleKey,
  "trkn" : tags.TrackNumberKey,
  "disk" : tags.DiscNumberKey,
  "tracknumber" : tags.TrackNumberKey,
  "TRACKNUMBER" : tags.TrackNumberKey,
  "DISKNUMBER" : tags.DiscNumberKey,
  "TIT2" : tags.TitleKey,
  "TPE1" : tags.ArtistKey,
  "TALB" : tags.AlbumKey,
}

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
    song := tags.GetTagsFromFile(path)
    if song == nil || len(song) == 0 {
      return nil
    }
    setPaths(song, path)
    song[tags.FlagsKey] = tags.EncodeFlag
    translateKeys(song)
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

// Replace keys with standard names.
func translateKeys(song tags.TagMap) {
  for k, v := range song {
    if trans, present := keyTranslations[k]; present {
      delete(song, k)
      song[trans] = v
    }
  }
  // Check for the track number.  If it doesn't exist, see if it has the TRCK tag, which
  // is track number / track total and get the track number from that.
  if _, present := song[tags.TrackNumberKey]; !present {
    if tntt, tnttPresent := song["TRCK"]; tnttPresent {
      s := strings.Split(tntt, "/")
      song[tags.TrackNumberKey] = s[0]
    } else {
      log.Printf("Can't get track number for '%s'\n", song[tags.RelativePathKey])
    }
  }
  // Check for the disc number.  If it doesn't exist, see if it has the TPOS tag, which
  // is disc number / disc total and get the disc number from that.  If that doesn't
  // exist, assume disc 1.
  if _, present := song[tags.DiscNumberKey]; !present {
    if dndt, dndtPresent := song["TPOS"]; dndtPresent {
      s := strings.Split(dndt, "/")
      song[tags.DiscNumberKey] = s[0]
    } else {
      song[tags.DiscNumberKey] = "1"
    }
  }
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

func songMapsToArtistMap(songMaps tags.TagMapSlice) map[string]common.Artist {
  artists := make(map[string]common.Artist)
  // Build a map of artists.
  for _, sm := range(songMaps) {
    name := sm[tags.ArtistKey]
    if _, present := artists[name]; !present {
      var artist common.Artist
      artist.Name = name
      artist.SortName = sm[tags.ArtistSortKey]
      artist.Albums = make(map[string]*common.Album)
      artists[name] = artist
    }
  }
  fmt.Printf("Found %d artists.\n", len(artists))
  // Build the maps of albums.
  numAlbums := 0
  for _, sm := range(songMaps) {
    name := sm[tags.ArtistKey]
    artist := artists[name]
    albumTitle := sm[tags.AlbumKey]
    if _, present := artist.Albums[albumTitle]; !present {
      album := new(common.Album)
      album.Title = albumTitle
      album.SortTitle = sm[tags.AlbumSortKey]
      album.Songs = make([]*common.Song, 0, 100)
      artist.Albums[albumTitle] = album
      numAlbums++
    }
  }
  fmt.Printf("Found %d albums.\n", numAlbums)
  // Build the lists of songs.  Note that we assume the songMap slice is sorted.
  for _, sm := range(songMaps) {
    name := sm[tags.ArtistKey]
    artist := artists[name]
    albumTitle := sm[tags.AlbumKey]
    album := artist.Albums[albumTitle]
    song := new(common.Song)
    song.Title = sm[tags.TitleKey]
    song.TrackNumber = btu.Atoi(sm[tags.TrackNumberKey])
    song.DiscNumber = btu.Atoi(sm[tags.DiscNumberKey])
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
      common.SortSongSlice(album.Songs)
    }
  }
  return artists
}

func artistMapToSongMaps(artistMap map[string]common.Artist) tags.TagMapSlice {
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
