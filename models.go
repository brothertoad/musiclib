package main

type ArtistModel struct {
  Id int `json:"id"`
  Name string `json:"name"`
}

type ArtistsResponse struct {
  Artists []ArtistModel `json:"artists"`
}
