package command

import (
  "fmt"
  "log"
  "io/ioutil"
  "github.com/urfave/cli/v2"
  "gopkg.in/yaml.v3"
)

var config struct {
  MusicDir string `yaml:"musicDir"`
  EncodedDir string `yaml:"encodedDir"`
  Mp3Dir string `yaml:"mp3Dir"`
  EncodeCommand interface{} `yaml:"encodeCommand"`
}

func LoadConfig(c *cli.Context) error {
  path := c.String("config")
  if !fileExists(path) {
    log.Fatalf("Config file '%s' does not exist.\n", path)
  }
  fmt.Printf("Loading configuration from %s...\n", path)
  b, err := ioutil.ReadFile(path)
  checkError(err)
  err = yaml.Unmarshal(b, &config)
  checkError(err)
  return nil
}
