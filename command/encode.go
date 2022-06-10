package command

import (
  "fmt"
  "io/ioutil"
  "log"
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

  // Before beginning encoding, make a clone of the encode command, and find out
  // which entries represent the input and output.
  args := make([]string, len(config.EncodeCommand))
  copy(args, config.EncodeCommand)
  inputIndex, outputIndex := -1, -1
  for j, arg := range(args) {
    if arg == "$INPUT" {
      inputIndex = j
    } else if arg == "$OUTPUT" {
      outputIndex = j
    }
  }
  if inputIndex < 0 || outputIndex < 0 {
    log.Fatalf("In encode command, missing input (%d) and/or output (%d)\n", inputIndex, outputIndex)
  }

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
      encodeSong(song, args, inputIndex, outputIndex)
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

func encodeSong(song common.Song, args []string, inputIndex int, outputIndex int) {
  fmt.Printf("Encoding %s...\n", song.RelativePath)
  inputPath := path.Join(config.MusicDir, song.RelativePath)
  outputPath := path.Join(config.EncodedDir, song.BasePath + song.EncodedExtension)
  args[inputIndex] = inputPath
  args[outputIndex] = outputPath
  fmt.Printf("%v\n", args)
}
