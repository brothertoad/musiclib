package command

import (
  "crypto/md5"
  "hash"
  "log"
  "io/ioutil"
  "github.com/urfave/cli/v2"
  "gopkg.in/yaml.v3"
)

type EncoderInfo struct {
  Extension string `yaml:"extension"`
  Directory string `yaml:"dir"`
  IncludeOtherEncodings string `yaml:"includeOtherEncodings"`
  Commands []string `yaml:"commands"`
  inputIndex int
  outputIndex int
  includeOthers bool
}

// Configuration.
var config struct {
  MusicDir string `yaml:"musicDir"`
  Mp3Dir string `yaml:"mp3Dir"`
  Encoders []EncoderInfo `yaml:"encoders"`
  DbUrl string `yaml:"dbUrl"`
}

// Other global data.
var verbose bool
var musicDirLength int
var hasher hash.Hash

func Init(c *cli.Context) error {
  // Load the config file.
  path := c.String("config")
  if !fileExists(path) {
    log.Fatalf("Config file '%s' does not exist.\n", path)
  }
  b, err := ioutil.ReadFile(path)
  checkError(err)
  err = yaml.Unmarshal(b, &config)
  checkError(err)
  // Verify our music directory is valid.
  if len(config.MusicDir) == 0 {
    log.Fatalln("No top level directory specified in configuration.")
  }
  dirMustExist(config.MusicDir)
  // Initialize the other global data.
  verbose = c.Bool("verbose")
  musicDirLength = len(config.MusicDir)
  hasher = md5.New()
  return nil
}
