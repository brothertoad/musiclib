package main

import (
  "os"
  "time"
  "github.com/urfave/cli/v2"
  "github.com/brothertoad/musiclib/command"
)

// Program for maintaining the music library database.  First argument is a command.
// Code for commands is in command directory.

func main() {
  app := &cli.App {
    Name: "musiclib",
    Compiled: time.Now(),
    Usage: "maintain musiclib database",
    Commands: []*cli.Command {
      &command.CreateCommand,
    },
  }
  app.Run(os.Args)
}
