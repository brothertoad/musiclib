package command

import (
  "fmt"
  "io/fs"
  "io/ioutil"
  "log"
  "path/filepath"
  "strings"
  "github.com/urfave/cli/v2"
  "gopkg.in/yaml.v3"
  "github.com/brothertoad/musiclib/common"
  "github.com/brothertoad/musiclib/tags"
)

const yamlFlag = "yaml"

var CreateCommand = cli.Command {
  Name: "create",
  Usage: "create (or recreate) the database",
  Action: doCreate,
  Flags: []cli.Flag {
    &cli.StringFlag {Name: yamlFlag},
  },
}

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

// In addition to being required, these are the only keys we save.
var requiredKeys = []string {
  common.TitleKey, common.ArtistKey, common.AlbumKey, common.TrackNumberKey, common.DiscNumberKey,
  common.ArtistSortKey, common.AlbumSortKey, common.PathKey, common.MimeKey, common.ExtensionKey,
  common.FlagsKey, common.DurationKey,
}

var songs []common.Song

func doCreate(c *cli.Context) error {
  fmt.Printf("Creating database from directory %s...\n", config.MusicDir)
  if len(config.MusicDir) == 0 {
    log.Fatalln("No top level directory specified in configuration.")
  }
  dirMustExist(config.MusicDir)
  songs = make([]common.Song, 0, 5000)
  filepath.WalkDir(config.MusicDir, loadFile)
  if len(c.String(yamlFlag)) > 0 {
    fmt.Printf("Saving yaml in '%s'\n", c.String(yamlFlag))
    data, err := yaml.Marshal(&songs)
    checkError(err)
    err = ioutil.WriteFile(c.String(yamlFlag), data, 0644)
    checkError(err)
  }
  fmt.Printf("Found %d songs.\n", len(songs))
  return nil
}

func loadFile(path string, de fs.DirEntry, err error) error {
  if de.IsDir() {
    return nil
  }
  song := tags.GetTagsFromFile(path)
  if song == nil || len(song) == 0 {
    return nil
  }
  song[common.PathKey] = path
  song[common.FlagsKey] = common.EncodeFlag
  translateKeys(song)
  addSortKeys(song)
  checkForMissingKeys(song)
  songs = append(songs, filterKeys(song))
  return nil
}

// Replace keys with standard names.
func translateKeys(song common.Song) {
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
      log.Printf("Can't get track number for '%s'\n", song[common.PathKey])
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

func addSortKeys(song common.Song) {
  addSortKey(song, common.ArtistKey, common.ArtistSortKey)
  addSortKey(song, common.AlbumKey, common.AlbumSortKey)
}

func addSortKey(song common.Song, pureKey string, sortKey string) {
  if _, present := song[sortKey]; !present {
    if vp, purePresent := song[pureKey]; purePresent {
      song[sortKey] = getSortValue(vp)
    } else {
      // If we don't have the pure key, we've got problems.
      log.Fatalf("no key '%s' for '%s'\n", pureKey, song[common.PathKey])
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

func checkForMissingKeys(song common.Song) {
  for _, k := range(requiredKeys) {
    if _, present := song[k]; !present {
      fmt.Printf("%+v is missing %s\n", song, k)
    }
  }
}

func filterKeys(song common.Song) common.Song {
  filtered := make(common.Song)
  for _, k := range(requiredKeys) {
    filtered[k] = song[k]
  }
  return filtered;
}
