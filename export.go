package main

import (
  "fmt"
  _ "io"
  _ "io/ioutil"
  _ "log"
  "maps"
  _ "os"
  _ "os/exec"
  _ "path"
  _ "path/filepath"
  "slices"
  "sort"
  "github.com/urfave/cli/v2"
  _ "github.com/brothertoad/btu"
)

var exportCommand = cli.Command {
  Name: "export",
  Usage: "export the database as text",
  Action: doExport,
}

func doExport(c *cli.Context) error {
  db := getDbConnection()
  defer db.Close()
  artists := slices.Collect(maps.Values(readArtistMapFromDb(db)))
  // fmt.Printf("%d artists are candidates for exporting\n", len(artists))
  sort.Slice(artists, func(i, j int) bool {
    return artists[i].SortName < artists[j].SortName
  })
  for _, artist := range(artists) {
    // fmt.Printf("%s:\n", artist.Name)
    // Create a slice of albums, sorted by SortTitle.
    albums := slices.Collect(maps.Values(artist.Albums))
    sort.Slice(albums, func(i, j int) bool {
      return albums[i].SortTitle < albums[j].SortTitle
    })
    for _, album := range(albums) {
      // fmt.Printf("%s   %s\n", artist.Name, album.Title)
      // Create a slice of songs, sorted by DiscNum and TrackNum.
      songs := album.Songs
      sort.Slice(songs, func(i, j int) bool {
        if songs[i].DiscNumber != songs[j].DiscNumber {
          return songs[i].DiscNumber < songs[j].DiscNumber
        }
        return songs[i].TrackNumber < songs[j].TrackNumber
      })
      for _, song := range(songs) {
        fmt.Printf("%s   %s   %d  %d  %s\n", artist.Name, album.Title, song.DiscNumber, song.TrackNumber, song.Title)
      }
    }
  }
  return nil
}
