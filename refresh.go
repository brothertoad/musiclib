package main

import (
  "fmt"
  "time"
  "github.com/urfave/cli/v2"
  "github.com/brothertoad/tags"
)

var refreshCommand = cli.Command {
  Name: "refresh",
  Usage: "refresh the database",
  Action: doRefresh,
}

func doRefresh(c *cli.Context) error {
  db := getDbConnection()
  defer db.Close()
  t0 := time.Now()
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
  t1 := time.Now()
  elapsed := t1.Sub(t0) // elapsed is nanoseconds
  seconds := (elapsed + 500000000) / 1000000000
  min := seconds / 60
  sec := seconds - (min * 60)
  fmt.Printf("refresh took %d:%02d\n", min, sec)
  return nil
}

func songMapSliceToSizeAndTimeMap(s tags.TagMapSlice) map[string]tags.TagMap {
  md5Map := make(map[string]tags.TagMap, len(s))
  for _, songMap := range(s) {
    md5Map[songMap[tags.SizeAndTimeKey]] = songMap
  }
  return md5Map
}

func findMissing(src, dest map[string]tags.TagMap) map[string]tags.TagMap {
  list := make(map[string]tags.TagMap, len(src))
  for srcSizeAndTime, srcSongMap := range(src) {
    _, present := dest[srcSizeAndTime]
    if !present {
      list[srcSizeAndTime] = srcSongMap
    }
  }
  return list
}
