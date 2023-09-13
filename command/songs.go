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
  "github.com/brothertoad/musiclib/common"
  "github.com/brothertoad/musiclib/tags"
)

var keyTranslations = map[string]string {
  "\xa9nam": common.TitleKey,
  "\xa9ART" : common.ArtistKey,
  "\xa9alb" : common.AlbumKey,
  "soar" : common.ArtistSortKey,
  "soal" : common.AlbumSortKey,
  "ALBUM" : common.AlbumKey,
  "ARTIST" : common.ArtistKey,
  "TITLE" : common.TitleKey,
  "trkn" : common.TrackNumberKey,
  "disk" : common.DiscNumberKey,
  "tracknumber" : common.TrackNumberKey,
  "TRACKNUMBER" : common.TrackNumberKey,
  "DISKNUMBER" : common.DiscNumberKey,
  "TIT2" : common.TitleKey,
  "TPE1" : common.ArtistKey,
  "TALB" : common.AlbumKey,
}

// In addition to being required, these are the only keys we save in the yaml file.
var requiredKeys = []string {
  common.TitleKey, common.ArtistKey, common.AlbumKey, common.TrackNumberKey, common.DiscNumberKey,
  common.ArtistSortKey, common.AlbumSortKey, common.RelativePathKey, common.BasePathKey,
  common.MimeKey, common.ExtensionKey, common.EncodedExtensionKey, common.IsEncodedKey,
  common.FlagsKey, common.DurationKey, common.Md5Key,
}

////////////////////////////////////////////////////////////////////////
//
// Logic for reading the library.
//
////////////////////////////////////////////////////////////////////////

func loadSongMapSliceFromMusicDir() common.SongMapSlice {
  songMaps := make(common.SongMapSlice, 0, 5000)
  filepath.WalkDir(config.MusicDir, func(path string, de fs.DirEntry, err error) error {
    if de.IsDir() {
      return nil
    }
    song := tags.GetTagsFromFile(path)
    if song == nil || len(song) == 0 {
      return nil
    }
    setPaths(song, path)
    song[common.FlagsKey] = common.EncodeFlag
    translateKeys(song)
    addSortKeys(song)
    addMd5Key(song)
    checkForMissingKeys(song)
    songMaps = append(songMaps, filterKeys(song))
    return nil
  })
  sort.Sort(songMaps)
  return songMaps
}

func setPaths(song common.SongMap, path string) {
  relativePath := path[musicDirLength:]
  song[common.RelativePathKey] = relativePath
  // Remove the extension to get the base path.
  pathLength := len(relativePath)
  extLength := len(song[common.ExtensionKey])
  song[common.BasePathKey] = relativePath[0:(pathLength-extLength)]
}

// Replace keys with standard names.
func translateKeys(song common.SongMap) {
  for k, v := range song {
    if trans, present := keyTranslations[k]; present {
      delete(song, k)
      song[trans] = v
    }
  }
  // Check for the track number.  If it doesn't exist, see if it has the TRCK tag, which
  // is track number / track total and get the track number from that.
  if _, present := song[common.TrackNumberKey]; !present {
    if tntt, tnttPresent := song["TRCK"]; tnttPresent {
      s := strings.Split(tntt, "/")
      song[common.TrackNumberKey] = s[0]
    } else {
      log.Printf("Can't get track number for '%s'\n", song[common.RelativePathKey])
    }
  }
  // Check for the disc number.  If it doesn't exist, see if it has the TPOS tag, which
  // is disc number / disc total and get the disc number from that.  If that doesn't
  // exist, assume disc 1.
  if _, present := song[common.DiscNumberKey]; !present {
    if dndt, dndtPresent := song["TPOS"]; dndtPresent {
      s := strings.Split(dndt, "/")
      song[common.DiscNumberKey] = s[0]
    } else {
      song[common.DiscNumberKey] = "1"
    }
  }
}

func addSortKeys(song common.SongMap) {
  addSortKey(song, common.ArtistKey, common.ArtistSortKey)
  addSortKey(song, common.AlbumKey, common.AlbumSortKey)
}

func addSortKey(song common.SongMap, pureKey string, sortKey string) {
  if _, present := song[sortKey]; !present {
    if vp, purePresent := song[pureKey]; purePresent {
      song[sortKey] = getSortValue(vp)
    } else {
      // If we don't have the pure key, we've got problems.
      log.Fatalf("no key '%s' for '%s'\n", pureKey, song[common.RelativePathKey])
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

func addMd5Key(song common.SongMap) {
  f, err := os.Open(path.Join(config.MusicDir,song[common.RelativePathKey]))
  btu.CheckError(err)
  defer f.Close()
  hasher.Reset()
  if _, err := io.Copy(hasher, f); err != nil {
    log.Fatalf("Error trying to compute md5sum of %s\n", song[common.RelativePathKey])
  }
  song[common.Md5Key] = hex.EncodeToString(hasher.Sum(nil))
}

func checkForMissingKeys(song common.SongMap) {
  for _, k := range(requiredKeys) {
    if _, present := song[k]; !present {
      fmt.Printf("%+v is missing %s\n", song, k)
    }
  }
}

func filterKeys(song common.SongMap) common.SongMap {
  filtered := make(common.SongMap)
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
    song := new(common.Song)
    song.Title = sm[common.TitleKey]
    song.TrackNumber = btu.Atoi(sm[common.TrackNumberKey])
    song.DiscNumber = btu.Atoi(sm[common.DiscNumberKey])
    song.Duration = sm[common.DurationKey]
    song.Mime = sm[common.MimeKey]
    song.Extension = sm[common.ExtensionKey]
    song.EncodedExtension = sm[common.EncodedExtensionKey]
    song.RelativePath = sm[common.RelativePathKey]
    song.BasePath = sm[common.BasePathKey]
    song.IsEncoded, _ = strconv.ParseBool(sm[common.IsEncodedKey])
    song.Flags = sm[common.FlagsKey]
    song.Md5 = sm[common.Md5Key]
    song.EncodedSourceMd5 = sm[common.EncodedSourceMd5Key]
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

func artistMapToSongMaps(artistMap map[string]common.Artist) common.SongMapSlice {
  songMaps := make(common.SongMapSlice, 0, 5000)
  for _, artist := range(artistMap) {
    for _, album := range(artist.Albums) {
      for _, song := range(album.Songs) {
        songMap := make(common.SongMap, 0)
        songMap[common.IdKey] = strconv.Itoa(song.Id)
        songMap[common.TitleKey] = song.Title
        songMap[common.ArtistKey] = artist.Name
        songMap[common.AlbumKey] = album.Title
        songMap[common.TrackNumberKey] = strconv.Itoa(song.TrackNumber)
        songMap[common.DiscNumberKey] = strconv.Itoa(song.DiscNumber)
        songMap[common.ArtistSortKey] = artist.SortName
        songMap[common.AlbumSortKey] = album.SortTitle
        songMap[common.RelativePathKey] = song.RelativePath
        songMap[common.BasePathKey] = song.BasePath
        songMap[common.MimeKey] = song.Mime
        songMap[common.ExtensionKey] = song.Extension
        songMap[common.EncodedExtensionKey] = song.EncodedExtension
        songMap[common.IsEncodedKey] = strconv.FormatBool(song.IsEncoded)
        songMap[common.FlagsKey] = song.Flags
        songMap[common.DurationKey] = song.Duration
        songMap[common.Md5Key] = song.Md5
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

func loadSongsFromYaml(path string) common.SongMapSlice {
  songMaps := make(common.SongMapSlice, 0, 5000)
  // logic came from https://zetcode.com/golang/yaml/
  yfile, err := ioutil.ReadFile(path)
  btu.CheckError(err)
  err2 := yaml.Unmarshal(yfile, &songMaps)
  btu.CheckError(err2)
  return songMaps
}

func saveSongsToYaml(path string, songMaps common.SongMapSlice) {
  fmt.Printf("Saving yaml in '%s'\n", path)
  data, err := yaml.Marshal(&songMaps)
  btu.CheckError(err)
  err = ioutil.WriteFile(path, data, 0644)
  btu.CheckError(err)
}
