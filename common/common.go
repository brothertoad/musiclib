package common

import (
  "sort"
  "strconv"
)

const FullPathKey = "fullPath"
const BasePathKey = "basePath"
const TitleKey = "title"
const ArtistKey = "artist"
const AlbumKey = "album"
const TrackNumberKey = "trackNumber"
const DiscNumberKey = "discNumber"
const ArtistSortKey = "artistSort"
const AlbumSortKey = "albumSort"
const DurationKey = "duration"
const MimeKey = "mime"
const ExtensionKey = "extension"
const EncodedExtensionKey = "encodedExtension"
const IsEncodedKey = "isEncoded"
const FlagsKey = "flags"
const Md5Key = "md5"
const EncodedSourceMd5Key = "encodedSourceMd5"

const EncodeFlag = "e"

type SongMap map[string]string
type SongMapSlice []SongMap

// functions for sorting a slice of SongMaps
func (s SongMapSlice) Len() int { return len(s) }
func (s SongMapSlice) Less(i, j int) bool {
  if s[i][ArtistSortKey] != s[j][ArtistSortKey] { return s[i][ArtistSortKey] < s[j][ArtistSortKey] }
  if s[i][AlbumSortKey] != s[j][AlbumSortKey] { return s[i][AlbumSortKey] < s[j][AlbumSortKey] }
  // convert disc numbers from strings to ints and check those
  di, _ := strconv.Atoi(s[i][DiscNumberKey])
  dj, _ := strconv.Atoi(s[j][DiscNumberKey])
  if di != dj { return di < dj }
  // convert track numbers from strings to ints and check those
  tni, _ := strconv.Atoi(s[i][TrackNumberKey])
  tnj, _ := strconv.Atoi(s[j][TrackNumberKey])
  return tni < tnj
}
func (s SongMapSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

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
  Serial int
  Title string
  TrackNumber int
  DiscNumber int
  Duration string
  Mime string
  Extension string
  EncodedExtension string
  FullPath string
  BasePath string
  IsEncoded bool
  Flags string
  Md5 string
  EncodedSourceMd5 string
  Sublibs string
}

type Album struct {
  Serial int
  Title string
  SortTitle string
  Songs []*Song
}

type Artist struct {
  Serial int
  Name string
  SortName string
  Albums map[string]*Album
}
