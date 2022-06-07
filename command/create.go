package command

import (
  "fmt"
  "io/ioutil"
  "sort"
  "github.com/urfave/cli/v2"
  "gopkg.in/yaml.v3"
  "github.com/brothertoad/musiclib/common"
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
  var songMaps common.SongMapSlice

  // If the load flag was specified, load from a file, rather than walking
  // through the entire music directory.
  if len(c.String(loadFlag)) > 0 {
    songMaps = loadSongsFromYaml(c.String(loadFlag))
  } else {
    songMaps = loadSongMapSliceFromMusicDir()
  }
  sort.Sort(songMaps)

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

func loadSongsFromYaml(path string) common.SongMapSlice {
  songMaps := make(common.SongMapSlice, 0, 5000)
  // logic came from https://zetcode.com/golang/yaml/
  yfile, err := ioutil.ReadFile(path)
  checkError(err)
  err2 := yaml.Unmarshal(yfile, &songMaps)
  checkError(err2)
  return songMaps
}
