package main

import (
  "sort"
)

// function for sorting a slice of Songs
func SortSongSlice(songs []*Song) {
  sort.Slice(songs, func(i, j int) bool {
    if songs[i].DiscNumber < songs[j].DiscNumber {
      return true
    }
    return songs[i].TrackNumber < songs[j].TrackNumber
  })
}

type Song struct {
  Id int
  Title string
  TrackNumber int
  DiscNumber int
  Duration string
  Mime string
  Extension string
  EncodedExtension string
  RelativePath string
  BasePath string
  IsEncoded bool
  Flags string
  State int
  SizeAndTime string
  Md5 string
  EncodedSource string
  Sublibs string
}

type Album struct {
  Id int
  Title string
  SortTitle string
  Songs []*Song
}

type Artist struct {
  Id int
  Name string
  SortName string
  Albums map[string]*Album
}
