package command

import (
  _ "crypto/md5"
  _ "encoding/hex"
  _ "fmt"
  _ "hash"
  _ "io"
  _ "io/fs"
  _ "io/ioutil"
  _ "log"
  _ "os"
  _ "path/filepath"
  _ "sort"
  _ "strings"
  "github.com/urfave/cli/v2"
  _ "gopkg.in/yaml.v3"
  _ "github.com/brothertoad/musiclib/common"
  _ "github.com/brothertoad/musiclib/tags"
)

var RefreshCommand = cli.Command {
  Name: "refresh",
  Usage: "refresh the database",
  Action: doRefresh,
}

func doRefresh(c *cli.Context) error {
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

func loadFile(path string, de fs.DirEntry, err error) error {
  if de.IsDir() {
    return nil
  }
  song := tags.GetTagsFromFile(path)
  if song == nil || len(song) == 0 {
    return nil
  }
  setPaths(song, path)
  song[common.FlagsKey] = common.EncodeFlag
  translateKeys(song)
  addSortKeys(song)
  addMd5Key(song)
  checkForMissingKeys(song)
  songMaps = append(songMaps, filterKeys(song))
  return nil
}

func setPaths(song common.SongMap, path string) {
  song[common.FullPathKey] = path
  // We will make the path relative to config.MusicDir, and remove the extension.
  pathLength := len(path)
  extLength := len(song[common.ExtensionKey])
  song[common.BasePathKey] = path[musicDirLength:(pathLength-extLength)]
}

// Replace keys with standard names.
func translateKeys(song common.SongMap) {
  for k, v := range song {
    if trans, present := keyTranslations[k]; present {
      delete(song, k)
      song[trans] = v
    }
  }
  // Check for the track number.  If it doesn't exist, see if it has the TRCK tag, which
  // is track number / track total and get the track number from that.
  if _, present := song[common.TrackNumberKey]; !present {
    if tntt, tnttPresent := song["TRCK"]; tnttPresent {
      s := strings.Split(tntt, "/")
      song[common.TrackNumberKey] = s[0]
    } else {
      log.Printf("Can't get track number for '%s'\n", song[common.FullPathKey])
    }
  }
  // Check for the disc number.  If it doesn't exist, see if it has the TPOS tag, which
  // is disc number / disc total and get the disc number from that.  If that doesn't
  // exist, assume disc 1.
  if _, present := song[common.DiscNumberKey]; !present {
    if dndt, dndtPresent := song["TPOS"]; dndtPresent {
      s := strings.Split(dndt, "/")
      song[common.DiscNumberKey] = s[0]
    } else {
      song[common.DiscNumberKey] = "1"
    }
  }
}

func addSortKeys(song common.SongMap) {
  addSortKey(song, common.ArtistKey, common.ArtistSortKey)
  addSortKey(song, common.AlbumKey, common.AlbumSortKey)
}

func addSortKey(song common.SongMap, pureKey string, sortKey string) {
  if _, present := song[sortKey]; !present {
    if vp, purePresent := song[pureKey]; purePresent {
      song[sortKey] = getSortValue(vp)
    } else {
      // If we don't have the pure key, we've got problems.
      log.Fatalf("no key '%s' for '%s'\n", pureKey, song[common.FullPathKey])
    }
  }
}

func getSortValue(pureValue string) string {
  if strings.HasPrefix(pureValue, "A ") {
    return pureValue[2:]
  }
  if strings.HasPrefix(pureValue, "a ") {
    return pureValue[2:]
  }
  if strings.HasPrefix(pureValue, "An ") {
    return pureValue[3:]
  }
  if strings.HasPrefix(pureValue, "an ") {
    return pureValue[3:]
  }
  if strings.HasPrefix(pureValue, "The ") {
    return pureValue[4:]
  }
  if strings.HasPrefix(pureValue, "the ") {
    return pureValue[4:]
  }
  return pureValue
}

func addMd5Key(song common.SongMap) {
  f, err := os.Open(song[common.FullPathKey])
  checkError(err)
  defer f.Close()
  hasher.Reset()
  if _, err := io.Copy(hasher, f); err != nil {
    log.Fatalf("Error trying to compute md5sum of %s\n", song[common.FullPathKey])
  }
  song[common.Md5Key] = hex.EncodeToString(hasher.Sum(nil))
}

func checkForMissingKeys(song common.SongMap) {
  for _, k := range(requiredKeys) {
    if _, present := song[k]; !present {
      fmt.Printf("%+v is missing %s\n", song, k)
    }
  }
}

func filterKeys(song common.SongMap) common.SongMap {
  filtered := make(common.SongMap)
  for _, k := range(requiredKeys) {
    filtered[k] = song[k]
  }
  return filtered;
}

*/