package command

import (
  "fmt"
  _ "io"
  _ "io/fs"
  _ "io/ioutil"
  _ "log"
  _ "os"
  _ "path/filepath"
  _ "sort"
  _ "strings"
  "github.com/urfave/cli/v2"
  _ "github.com/brothertoad/musiclib/common"
  _ "github.com/brothertoad/musiclib/tags"
)

var RefreshCommand = cli.Command {
  Name: "refresh",
  Usage: "refresh the database",
  Action: doRefresh,
}

func doRefresh(c *cli.Context) error {
  diskSongMaps := loadSongMapSliceFromMusicDir()
  dbSongMaps := artistMapToSongMaps(readArtistMapFromDb())
  fmt.Printf("%d songs on disk, %d songs in database\n", len(diskSongMaps), len(dbSongMaps))
  return nil
}
