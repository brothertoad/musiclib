package main

import (
  "os"
  "time"
  "github.com/urfave/cli/v2"
  "github.com/brothertoad/musiclib/command"
)

// Program for maintaining the music library database.  First non-option argument is a command.
// Code for commands is in command directory.

// Commands: create, update (alias is refresh), serve, encode, mp3 (perhaps part of encode?), sublib
// Perhaps allow sublib by extension, so don't need separate mp3 command.

func main() {
  app := &cli.App {
    Name: "musiclib",
    Compiled: time.Now(),
    Usage: "maintain musiclib database",
    Flags: []cli.Flag {
      &cli.StringFlag {Name: "config", Required: true, EnvVars: []string{"MUSICLIB_CONFIG"},},
    },
    Commands: []*cli.Command {
      &command.CreateCommand,
      &command.RefreshCommand,
    },
    Before: command.Init,
  }
  app.Run(os.Args)
}
