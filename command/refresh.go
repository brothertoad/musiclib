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
  "github.com/brothertoad/musiclib/common"
)

var RefreshCommand = cli.Command {
  Name: "refresh",
  Usage: "refresh the database",
  Action: doRefresh,
}

func doRefresh(c *cli.Context) error {
  db := getDbConnection()
  defer db.Close()
  diskSongMaps := loadSongMapSliceFromMusicDir()
  dbSongMaps := artistMapToSongMaps(readArtistMapFromDb(db))
  diskMd5s := songMapSliceToMd5Map(diskSongMaps)
  dbMd5s := songMapSliceToMd5Map(dbSongMaps)
  added := findMissing(diskMd5s, dbMd5s)
  deleted := findMissing(dbMd5s, diskMd5s)
  fmt.Printf("%d songs added, %d songs deleted\n", len(added), len(deleted))
  deleteSongsFromDb(db, deleted)
  addSongsToDb(db, added)
  deleteEmptyContainers(db)
  return nil
}

func songMapSliceToMd5Map(s common.SongMapSlice) map[string]common.SongMap {
  md5Map := make(map[string]common.SongMap, len(s))
  for _, songMap := range(s) {
    md5Map[songMap[common.Md5Key]] = songMap
  }
  return md5Map
}

func findMissing(src, dest map[string]common.SongMap) map[string]common.SongMap {
  list := make(map[string]common.SongMap, len(src))
  for srcMd5, srcSongMap := range(src) {
    _, present := dest[srcMd5]
    if !present {
      list[srcMd5] = srcSongMap
    }
  }
  return list
}
