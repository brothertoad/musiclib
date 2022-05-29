package command

import (
  "fmt"
  "io/fs"
  "log"
  "path/filepath"
  "strings"
  "github.com/urfave/cli/v2"
  "github.com/brothertoad/musiclib/common"
  "github.com/brothertoad/musiclib/tags"
)

var CreateCommand = cli.Command {
  Name: "create",
  Usage: "create (or recreate) the database",
  Action: doCreate,
}

// Perhaps use constants from commom for target keys.
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

var requiredKeys = []string {
  common.TitleKey, common.ArtistKey, common.AlbumKey, common.TrackNumberKey, common.DiscNumberKey,
}

var songs []common.Song

func doCreate(c *cli.Context) error {
  fmt.Printf("Creating database from directory %s...\n", config.MusicDir)
  if len(config.MusicDir) == 0 {
    log.Fatalln("No top level directory specified in configuration.")
  }
  dirMustExist(config.MusicDir)
  songs = make([]common.Song, 5000)
  filepath.WalkDir(config.MusicDir, loadFile)
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
  translateKeys(song)
  // Add sort keys.
  song[common.FlagsKey] = ""
  checkForMissingKeys(song)
  songs = append(songs, song)
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

}

func checkForMissingKeys(song common.Song) {
  for _, k := range(requiredKeys) {
    if _, present := song[k]; !present {
      fmt.Printf("%+v is missing %s\n", song, k)
    }
  }
}
