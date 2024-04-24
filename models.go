package main

type ArtistModel struct {
  Id int `json:"id"`
  Name string `json:"name"`
}

type AlbumModel struct {
  Id int `json:"id"`
  Title string `json:"title"`
}
