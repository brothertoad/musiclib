package command

import (
  "fmt"
  "github.com/urfave/cli/v2"
)

type Config struct {
  dir string
}

var config Config

func LoadConfig(c *cli.Context) error {
  fmt.Printf("Loading configuration from %s...\n", c.String("config"))
  return nil
}
