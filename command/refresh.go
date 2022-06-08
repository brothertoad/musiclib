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
  dbArtistMap := readArtistMapFromDb()
  fmt.Printf("%d songs on disk, %d artists in database\n", len(diskSongMaps), len(dbArtistMap))
  return nil
}

/*

var songMaps common.SongMapSlice
var musicDirLength int
var hasher hash.Hash

func doRefresh(c *cli.Context) error {
  fmt.Printf("Refreshing database from directory %s...\n", config.MusicDir)
  // save the length, as we need it to remove the prefix of each file
  musicDirLength = len(config.MusicDir)
  hasher = md5.New()
  songMaps = make(common.SongMapSlice, 0, 5000)

  filepath.WalkDir(config.MusicDir, loadFile)
  sort.Sort(songMaps)

  fmt.Printf("Found %d songs.\n", len(songMaps))
  addArtistMapToDb(songMapsToArtistMap(songMaps))
  return nil
}
*/
