package main

import (
  "os"
  "time"
  "github.com/urfave/cli/v2"
)

// Program for maintaining the music library database.  First non-option argument is a command.
// Code for commands is in command directory.

// Commands: create, refresh, serve, encode, mp3 (perhaps part of encode?), sublib, export (csv, pdf)
// Perhaps allow sublib by extension rather than flag, so don't need separate mp3 command.
// Or simply use rclone/rsync for mp3.

// TASKS: Add indexes to database, after populating it.  Use the log-level flag. Add a library
// column to albums, and change configuration to support multiple libraries (one of which is the
// default).  If library column is not used, then change database creation scripts to use a different
// role for each database.

func main() {
  app := &cli.App {
    Name: "musiclib",
    Compiled: time.Now(),
    Usage: "maintain musiclib database",
    Flags: []cli.Flag {
      &cli.StringFlag {Name: "config", Required: true, EnvVars: []string{"MUSICLIB_CONFIG"},},
      &cli.BoolFlag {Name: "verbose", Aliases: []string{"v"}},
      &cli.StringFlag {Name: "log-level"},
    },
    Commands: []*cli.Command {
      &CreateCommand,
      &RefreshCommand,
      &EncodeCommand,
      &serveCommand,
    },
    Before: Init,
  }
  app.Run(os.Args)
}
