package main

import (
  "fmt"
  "sort"
  "github.com/urfave/cli/v2"
  "github.com/brothertoad/tags"
)

const saveFlag = "save"
const loadFlag = "load"
const statsFlag = "stats"
const dryRunFlag = "dry-run"
const csvFlag = "csv"

var createCommand = cli.Command {
  Name: "create",
  Usage: "create (or recreate) the database",
  Action: doCreate,
  Flags: []cli.Flag {
    &cli.StringFlag {Name: saveFlag},
    &cli.StringFlag {Name: csvFlag},
    &cli.StringFlag {Name: loadFlag},
    &cli.BoolFlag {Name: statsFlag},
    &cli.BoolFlag {Name: dryRunFlag, Aliases: []string{"n"}},
  },
}

func doCreate(c *cli.Context) error {
  verbose := c.Bool(verboseFlag)
  if verbose {
    fmt.Printf("Creating database from directory %s...\n", config.MusicDir)
  }
  var songMaps tags.TagMapSlice

  // If the load flag was specified, load from a file, rather than walking
  // through the entire music directory.
  if len(c.String(loadFlag)) > 0 {
    songMaps = loadSongsFromYaml(c.String(loadFlag))
  } else {
    songMaps = loadSongMapSliceFromMusicDir()
  }
  sort.Sort(songMaps)

  // If the csv flag was specified, then save the songs as a CSV file.
  if len(c.String(csvFlag)) > 0 {
    saveSongsToCsv(c.String(csvFlag), songMaps)
  }

  // Save if the save flag was specified.
  if len(c.String(saveFlag)) > 0 {
    saveSongsToYaml(c.String(saveFlag), songMaps)
  }

  stats := c.Bool(statsFlag)
  if stats {
    fmt.Printf("Found %d songs.\n", len(songMaps))
  }
  if !c.Bool(dryRunFlag) {
    db := getDbConnection()
    defer db.Close()
    addArtistMapToDb(db, songMapsToArtistMap(songMaps, stats))
  }
  return nil
}
