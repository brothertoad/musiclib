package main

type ArtistModel struct {
  Id int `json:"id"`
  Name string `json:"name"`
  SortName string `json:"sortName"`
}

type ArtistsResponse struct {
  Artists []ArtistModel `json:"artists"`
}
