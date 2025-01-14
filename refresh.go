package main

import (
  "database/sql"
  "fmt"
  "time"
  "github.com/urfave/cli/v2"
  "github.com/brothertoad/btu"
  "github.com/brothertoad/tags"
)

var useMd5 = false

var refreshCommand = cli.Command {
  Name: "refresh",
	Flags: []cli.Flag {
	  &cli.BoolFlag {Name: "md5", Aliases: []string{"m"}, Value: false, Destination: &useMd5},
	},
  Usage: "refresh the database",
  Action: doRefresh,
}

func doRefresh(c *cli.Context) error {
  db := getDbConnection()
  defer db.Close()
  t0 := time.Now()
  if verbose {
	  fmt.Printf("About to load songs from disk %s\n", time.Now().Format(time.TimeOnly))
  }
  diskSongMaps := loadSongMapSliceFromMusicDir(useMd5)
  if verbose {
	  fmt.Printf("About to load songs from database %s\n", time.Now().Format(time.TimeOnly))
  }
  dbSongMaps := artistMapToSongMaps(readArtistMapFromDb(db))
  if verbose {
	  fmt.Printf("About to convert keys from disk songs %s\n", time.Now().Format(time.TimeOnly))
  }
  diskKeys := songMapSliceToSizeAndTimeMap(diskSongMaps)
  if verbose {
	  fmt.Printf("About to convert keys from database songs %s\n", time.Now().Format(time.TimeOnly))
  }
  dbKeys := songMapSliceToSizeAndTimeMap(dbSongMaps)
  if verbose {
	  fmt.Printf("About to calculate the number of songs that moved %s\n", time.Now().Format(time.TimeOnly))
  }
  numMoved := updatePaths(db, dbKeys, diskKeys)
  if numMoved > 0 {
    fmt.Printf("%d songs moved\n", numMoved)
  }
  if verbose {
	  fmt.Printf("About to find added %s\n", time.Now().Format(time.TimeOnly))
  }
  added := findMissing(diskKeys, dbKeys)
  if verbose {
	  fmt.Printf("About to find deleted %s\n", time.Now().Format(time.TimeOnly))
  }
  deleted := findMissing(dbKeys, diskKeys)
  fmt.Printf("%d songs added, %d songs deleted\n", len(added), len(deleted))
  if verbose {
	  fmt.Printf("About to delete songs from database %s\n", time.Now().Format(time.TimeOnly))
  }
  deleteSongsFromDb(db, deleted)
  if verbose {
	  fmt.Printf("About to add songs to database %s\n", time.Now().Format(time.TimeOnly))
  }
  addSongsToDb(db, added)
  if verbose {
	  fmt.Printf("About to delete empty containers in database %s\n", time.Now().Format(time.TimeOnly))
  }
  deleteEmptyContainers(db)
  if verbose {
	  fmt.Printf("Done deleting empty containers in database %s\n", time.Now().Format(time.TimeOnly))
  }
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

func updatePaths(db *sql.DB, stale, fresh map[string]tags.TagMap) int {
    total := 0
    for k, v := range fresh {
      ov, found := stale[k]
      if !found {
        continue // ignore fresh songs that aren't in stale
      }
      freshPath := v[tags.RelativePathKey]
      stalePath := ov[tags.RelativePathKey]
      if freshPath != stalePath {
        // Note that since the fresh map was read from disk, there are no ids.
        id := btu.Atoi2(ov[tags.IdKey], "Can't convert '%s' to a song id", ov[tags.IdKey])
        updateSongPaths(db, id, v)
        total++
      }
    }
    return total
}
