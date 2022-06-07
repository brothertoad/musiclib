package command

import (
  _ "encoding/hex"
  "fmt"
  _ "hash"
  _ "io"
  "io/fs"
  "io/ioutil"
  _ "log"
  _ "os"
  "path/filepath"
  "sort"
  _ "strings"
  "github.com/urfave/cli/v2"
  "gopkg.in/yaml.v3"
  "github.com/brothertoad/musiclib/common"
  "github.com/brothertoad/musiclib/tags"
)

const saveFlag = "save"
const loadFlag = "load"

var CreateCommand = cli.Command {
  Name: "create",
  Usage: "create (or recreate) the database",
  Action: doCreate,
  Flags: []cli.Flag {
    &cli.StringFlag {Name: saveFlag},
    &cli.StringFlag {Name: loadFlag},
  },
}

func doCreate(c *cli.Context) error {
  fmt.Printf("Creating database from directory %s...\n", config.MusicDir)
  // save the length, as we need it to remove the prefix of each file
  songMaps := make(common.SongMapSlice, 0, 5000)

  // If the load flag was specified, load from a file, rather than walking
  // through the entire music directory.
  if len(c.String(loadFlag)) > 0 {
    loadSongs(c.String(loadFlag), songMaps)
  } else {
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
  }

  // Save if the save flag was specified.
  if len(c.String(saveFlag)) > 0 {
    fmt.Printf("Saving yaml in '%s'\n", c.String(saveFlag))
    data, err := yaml.Marshal(&songMaps)
    checkError(err)
    err = ioutil.WriteFile(c.String(saveFlag), data, 0644)
    checkError(err)
  }

  fmt.Printf("Found %d songs.\n", len(songMaps))
  addArtistMapToDb(songMapsToArtistMap(songMaps))
  return nil
}

func loadSongs(path string, songMaps common.SongMapSlice) {
  // logic came from https://zetcode.com/golang/yaml/
  yfile, err := ioutil.ReadFile(path)
  checkError(err)
  err2 := yaml.Unmarshal(yfile, &songMaps)
  checkError(err2)
}
