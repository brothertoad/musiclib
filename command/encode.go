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
)

var EncodeCommand = cli.Command {
  Name: "encode",
  Usage: "encode the database",
  Action: doEncode,
}

func doEncode(c *cli.Context) error {
  db := getDbConnection()
  defer db.Close()
  songMaps := artistMapToSongMaps(readArtistMapFromDb(db))
  fmt.Printf("%d songs are candidates for encoding\n", len(songMaps))
  return nil
}
