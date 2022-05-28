package command

import (
  "fmt"
  "io/fs"
  "log"
  "os"
  "path/filepath"
  "github.com/urfave/cli/v2"
  "github.com/brothertoad/musiclib/common"
  "github.com/brothertoad/musiclib/tags"
)

var CreateCommand = cli.Command {
  Name: "create",
  Usage: "create (or recreate) the database",
  Action: doCreate,
}

var keyTranslations = map[string]string {
  "xyz": "abc",
}

var songs []common.Song

func doCreate(c *cli.Context) error {
  fmt.Printf("Creating database from directory %s...\n", config.MusicDir)
  if len(config.MusicDir) == 0 {
    log.Fatalln("No top level directory specified in configuration.")
  }
  if _, err := os.Stat(config.MusicDir); os.IsNotExist(err) {
  	log.Fatalf("Top level directory '%s' does not exist.\n", config.MusicDir)
  }
  songs = make([]common.Song, 5000)
  filepath.WalkDir(config.MusicDir, loadFile)
  return nil
}

func loadFile(path string, de fs.DirEntry, err error) error {
  if de.IsDir() {
    return nil
  }
  m := tags.GetTagsFromFile(path)
  if m == nil || len(m) == 0 {
    return nil
  }
  // Standardize keys.
  // Add sort keys.
  // Add flags.
  songs = append(songs, m)
  return nil
}
