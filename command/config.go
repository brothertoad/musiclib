package command

import (
  "fmt"
  "github.com/urfave/cli/v2"
)

var config struct {
  dir string
}

func LoadConfig(c *cli.Context) error {
  fmt.Printf("Loading configuration from %s...\n", c.String("config"))
  return nil
}
