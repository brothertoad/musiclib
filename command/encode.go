package command

import (
  "fmt"
  "io"
  "io/ioutil"
  "log"
  "os"
  "os/exec"
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

type extendedEncoderInfo struct {
  EncoderInfo
  inputIndex int
  outputIndex int
}

func doEncode(c *cli.Context) error {
  db := getDbConnection()
  defer db.Close()
  songs := readSongListFromDb(db)
  fmt.Printf("%d songs are candidates for encoding\n", len(songs))

  findEncoderIndices()

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

func findEncoderIndices() {
  for i, encoder := range(config.Encoders) {
    inputIndex := -1
    outputIndex := -1
    for j, arg := range(encoder.Commands) {
      if arg == "$INPUT" {
        inputIndex = j
      } else if arg == "$OUTPUT" {
        outputIndex = j
      }
    }
    if inputIndex < 0 || outputIndex < 0 {
      log.Fatalf("Missing either $INPUT or $OUTPUT for encoder %+v\n", encoder)
    }
    config.Encoders[i].inputIndex = inputIndex
    config.Encoders[i].outputIndex = outputIndex
  }
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
  inputPath := path.Join(config.MusicDir, song.RelativePath)
  for _, encoder := range(config.Encoders) {
    outputPath := path.Join(config.EncodedDir, song.BasePath + encoder.Extension)
    err := os.MkdirAll(filepath.Dir(outputPath), 0775)
    checkError(err)
    encoder.Commands[encoder.inputIndex] = inputPath
    encoder.Commands[encoder.outputIndex] = outputPath
    cmd := exec.Command(encoder.Commands[0], encoder.Commands[1:]...)
    stderr, err := cmd.StderrPipe()
    checkError(err)
    err = cmd.Start()
    checkError(err)
    _, _ = io.ReadAll(stderr)
    err = cmd.Wait()
    checkError(err)
  }
}
