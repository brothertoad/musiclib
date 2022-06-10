package command

import (
  "fmt"
  "io/ioutil"
  "os"
  "path"
  "path/filepath"
  "github.com/urfave/cli/v2"
  "github.com/brothertoad/musiclib/common"
)

var EncodeCommand = cli.Command {
  Name: "encode",
  Usage: "encode the database",
  Action: doEncode,
}

func doEncode(c *cli.Context) error {
  db := getDbConnection()
  defer db.Close()
  songs := readSongListFromDb(db)
  fmt.Printf("%d songs are candidates for encoding\n", len(songs))
  for _, song := range(songs) {
    // Regardless of whether or not the source file is already encoded,
    // if there is an encodedSourceMd5 and it matches the current Md5,
    // we don't need to do anything.
    if song.Md5 == song.EncodedSourceMd5 {
      continue
    }
    if song.IsEncoded {
      copySong(song)
    } else {
      encodeSong(song)
    }
    song.EncodedSourceMd5 = song.Md5
    updateSongEncodedSourceMd5(db, song)
  }
  return nil
}

func copySong(song common.Song) {
  fmt.Printf("Copying %s...\n", song.RelativePath)
  src := path.Join(config.MusicDir, song.RelativePath)
  dest := path.Join(config.EncodedDir, song.BasePath + song.EncodedExtension)
  err := os.MkdirAll(filepath.Dir(dest), 0775)
  checkError(err)
  bytes, err := ioutil.ReadFile(src)
  checkError(err)
  err = ioutil.WriteFile(dest, bytes, 0644)
}

func encodeSong(song common.Song) {
  fmt.Printf("Encoding %s...\n", song.RelativePath)
}
