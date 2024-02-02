package command

import (
  "fmt"
  "sort"
  "github.com/urfave/cli/v2"
  "github.com/brothertoad/tags"
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
  var songMaps tags.TagMapSlice

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
    saveSongsToYaml(c.String(saveFlag), songMaps)
  }

  fmt.Printf("Found %d songs.\n", len(songMaps))
  db := getDbConnection()
  defer db.Close()
  addArtistMapToDb(db, songMapsToArtistMap(songMaps))
  return nil
}
