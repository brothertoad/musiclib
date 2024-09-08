package main

type ArtistModel struct {
  Id int `json:"id"`
  Name string `json:"name"`
}

type AlbumModel struct {
  Id int `json:"id"`
  Title string `json:"title"`
}

type SongModel struct {
  Id int `json:"id"`
  TrackNum int `json:"trackNum"`
  Title string `json:"title"`
  Album string `json:"album"`
  Artist string `json:"artist"`
}

type UpdateSongStatesModel struct {
  State int `json:"state"`
  SongIds []int `json:"songIds"`
}
