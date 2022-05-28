package command

import (
  "fmt"
  "io/fs"
  "path/filepath"
  "github.com/urfave/cli/v2"
  "github.com/brothertoad/musiclib/tags"
)

const dirFlag string = "dir"

var CreateCommand = cli.Command {
  Name: "create",
  Usage: "create (or recreate) the database",
  Action: create,
  Flags: []cli.Flag {
    &cli.StringFlag {Name: dirFlag, Value: "", Usage: "top level directory to search", Required: true,},
  },
}

func create(c *cli.Context) error {
  fmt.Printf("Creating database from directory %s...\n", c.String(dirFlag))
  filepath.WalkDir(c.String(dirFlag), loadFile)
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
