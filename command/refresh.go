package command

import (
  "fmt"
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
  diskKeys := songMapSliceToSizeAndTimeMap(diskSongMaps)
  dbKeys := songMapSliceToSizeAndTimeMap(dbSongMaps)
  added := findMissing(diskKeys, dbKeys)
  deleted := findMissing(dbKeys, diskKeys)
  fmt.Printf("%d songs added, %d songs deleted\n", len(added), len(deleted))
  deleteSongsFromDb(db, deleted)
  addSongsToDb(db, added)
  deleteEmptyContainers(db)
  return nil
}

func songMapSliceToSizeAndTimeMap(s common.SongMapSlice) map[string]common.SongMap {
  md5Map := make(map[string]common.SongMap, len(s))
  for _, songMap := range(s) {
    md5Map[songMap[common.SizeAndTimeKey]] = songMap
  }
  return md5Map
}

func findMissing(src, dest map[string]common.SongMap) map[string]common.SongMap {
  list := make(map[string]common.SongMap, len(src))
  for srcSizeAndTime, srcSongMap := range(src) {
    _, present := dest[srcSizeAndTime]
    if !present {
      list[srcSizeAndTime] = srcSongMap
    }
  }
  return list
}
