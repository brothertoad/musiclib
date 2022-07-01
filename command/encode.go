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
  "strconv"
  "github.com/urfave/cli/v2"
  "github.com/brothertoad/btu"
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

  validateEncoders()

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

// Get the input and output indices for each encoder, and set the output
// directory if it was not explicitly specified.
func validateEncoders() {
  for i, encoder := range(config.Encoders) {
    if encoder.Extension == "" {
      log.Fatalf("Encoder %v does not have an extension\n", encoder)
    }
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
    if encoder.Directory == "" {
      config.Encoders[i].Directory = config.MusicDir + "-" + config.Encoders[i].Extension
    }
    // Set includeOthers based on string provided in yaml file.  Note that the default
    // is true, which is why we can't just use a bool in the yaml file.
    if encoder.IncludeOtherEncodings == "" {
      config.Encoders[i].includeOthers = true
    } else {
      includeOthers, err := strconv.ParseBool(encoder.IncludeOtherEncodings)
      btu.CheckError(err)
      config.Encoders[i].includeOthers = includeOthers
    }
  }
}

func copySong(song common.Song) {
  src := path.Join(config.MusicDir, song.RelativePath)
  for _, encoder := range(config.Encoders) {
    // We only copy the file if the extension is the same as the encoder,
    // or if the encoder is configured to include other encodings.
    if song.Extension == encoder.Extension || encoder.includeOthers {
      fmt.Printf("Copying %s...\n", song.RelativePath)
      dest := path.Join(encoder.Directory, song.BasePath + song.EncodedExtension)
      err := os.MkdirAll(filepath.Dir(dest), 0775)
      btu.CheckError(err)
      bytes, err := ioutil.ReadFile(src)
      btu.CheckError(err)
      err = ioutil.WriteFile(dest, bytes, 0644)
    }
  }
}

func encodeSong(song common.Song) {
  fmt.Printf("Encoding %s...\n", song.RelativePath)
  inputPath := path.Join(config.MusicDir, song.RelativePath)
  for _, encoder := range(config.Encoders) {
    outputPath := path.Join(encoder.Directory, song.BasePath + encoder.Extension)
    err := os.MkdirAll(filepath.Dir(outputPath), 0775)
    btu.CheckError(err)
    encoder.Commands[encoder.inputIndex] = inputPath
    encoder.Commands[encoder.outputIndex] = outputPath
    cmd := exec.Command(encoder.Commands[0], encoder.Commands[1:]...)
    stderr, err := cmd.StderrPipe()
    btu.CheckError(err)
    err = cmd.Start()
    btu.CheckError(err)
    _, _ = io.ReadAll(stderr)
    err = cmd.Wait()
    btu.CheckError(err)
  }
}
