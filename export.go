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
  "strings"
  "github.com/urfave/cli/v2"
  _ "github.com/brothertoad/btu"
)

var exportCommand = cli.Command {
  Name: "export",
  Usage: "export the database as text",
  Action: doExport,
}

type songInfo struct {
  artist, album string
  disc, track int
  title string
}

func doExport(c *cli.Context) error {
  db := getDbConnection()
  defer db.Close()

  // Create a slice of songInfo, and put all the songs in it, sorted.
  // This allows us to find the longest artist, album and title, so
  // we can print out the song list more cleanly.
  songInfoList := make([]songInfo, 0, 10000)

  artists := slices.Collect(maps.Values(readArtistMapFromDb(db)))
  // fmt.Printf("%d artists are candidates for exporting\n", len(artists))
  sort.Slice(artists, func(i, j int) bool {
    return strings.ToUpper(artists[i].SortName) < strings.ToUpper(artists[j].SortName)
  })
  for _, artist := range(artists) {
    // fmt.Printf("%s:\n", artist.Name)
    // Create a slice of albums, sorted by SortTitle.
    albums := slices.Collect(maps.Values(artist.Albums))
    sort.Slice(albums, func(i, j int) bool {
      return strings.ToUpper(albums[i].SortTitle) < strings.ToUpper(albums[j].SortTitle)
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
        var info songInfo
        info.artist = artist.Name
        info.album = album.Title
        info.disc = song.DiscNumber
        info.track = song.TrackNumber
        info.title = song.Title
        songInfoList = append(songInfoList, info)
      }
    }
  }
  for _, info := range(songInfoList) {
    fmt.Printf("%s   %s   %d  %d  %s\n", info.artist, info.album, info.disc, info.track, info.title)
  }
  return nil
}
