package command

import (
  "fmt"
  "io/fs"
  "log"
  "os"
  "path/filepath"
  "github.com/urfave/cli/v2"
  "github.com/brothertoad/musiclib/tags"
)

var CreateCommand = cli.Command {
  Name: "create",
  Usage: "create (or recreate) the database",
  Action: doCreate,
}

func doCreate(c *cli.Context) error {
  fmt.Printf("Creating database from directory %s...\n", config.dir)
  if len(config.dir) == 0 {
    log.Fatalln("No top level directory specified in configuration.")
  }
  if _, err := os.Stat(config.dir); os.IsNotExist(err) {
  	log.Fatalf("Top level directory '%s' does not exist.\n", config.dir)
  }
  filepath.WalkDir(config.dir, loadFile)
  return nil
}

func loadFile(path string, de fs.DirEntry, err error) error {
  if de.IsDir() {
    return nil
  }
  // fmt.Printf("Loading %s\n", path)
  m := tags.GetTagsFromFile(path)
  if m == nil || len(m) == 0 {
    fmt.Printf("Got nil or empty map for %s\n", path)
  }
  return nil
}
