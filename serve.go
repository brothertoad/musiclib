package main

import (
  "database/sql"
  "fmt"
  "net/http"
  _ "sort"
  "strconv"
  "github.com/labstack/echo/v4"
  "github.com/labstack/echo/v4/middleware"
  "github.com/urfave/cli/v2"
)

var serveCommand = cli.Command {
  Name: "serve",
  Usage: "run as a REST service",
  Flags: []cli.Flag {
    &cli.IntFlag {Name: "port", Aliases: []string{"p"}, Value: 9904},
  },
  Action: doServe,
}

func doServe(c *cli.Context) error {
  port := c.Int("port")
  db := getDbConnection()
	defer db.Close()

	e := echo.New()
  e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
    Format: "${time_rfc3339} ${method} uri=${uri} status=${status} error=${error}\n",
  }))
  e.Use(middleware.CORS()) // allow all requests

  e.GET("/artists/:state", func(c echo.Context) error {
		return getArtists(c, db)
	})
  e.GET("/albums/:artistId/:state", func(c echo.Context) error {
		return getAlbums(c, db)
	})
  e.GET("/songs/:albumId/:state", func(c echo.Context) error {
		return getSongs(c, db)
	})
  e.POST("/updatesongs", func(c echo.Context) error {
		return updateSongStates(c, db)
	})

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
  return nil
}

func getArtists(c echo.Context, db *sql.DB) error {
  stateString := c.Param("state")
  state, err := strconv.Atoi(stateString)
  if err != nil {
    return c.String(http.StatusBadRequest, fmt.Sprintf("Can't convert state '%s' to a number\n", stateString))
  }
  artists, err := loadArtists(db, state)
  if err != nil {
    return c.String(http.StatusBadRequest, "error loading artists by state\n")
  }
  return c.JSON(http.StatusOK, artists)
}

func getAlbums(c echo.Context, db *sql.DB) error {
  artistString := c.Param("artistId")
  artistId, err := strconv.Atoi(artistString)
  if err != nil {
    return c.String(http.StatusBadRequest, fmt.Sprintf("Can't convert artistId '%s' to a number\n", artistString))
  }
  stateString := c.Param("state")
  state, err := strconv.Atoi(stateString)
  if err != nil {
    return c.String(http.StatusBadRequest, fmt.Sprintf("Can't convert state '%s' to a number\n", stateString))
  }
  artists, err := loadAlbums(db, artistId, state)
  if err != nil {
    return c.String(http.StatusBadRequest, "error loading albums\n")
  }
  return c.JSON(http.StatusOK, artists)
}

func getSongs(c echo.Context, db *sql.DB) error {
  albumString := c.Param("albumId")
  albumId, err := strconv.Atoi(albumString)
  if err != nil {
    return c.String(http.StatusBadRequest, fmt.Sprintf("Can't convert albumId '%s' to a number\n", albumString))
  }
  stateString := c.Param("state")
  state, err := strconv.Atoi(stateString)
  if err != nil {
    return c.String(http.StatusBadRequest, fmt.Sprintf("Can't convert state '%s' to a number\n", stateString))
  }
  songs, err := loadSongs(db, albumId, state)
  if err != nil {
    return c.String(http.StatusBadRequest, fmt.Sprintf("error loading songs: %s\n", err.Error()))
  }
  return c.JSON(http.StatusOK, songs)
}

func updateSongStates(c echo.Context, db *sql.DB) error {
  updateModel := new(UpdateSongStatesModel)
  if err := c.Bind(updateModel); err != nil {
    return c.String(http.StatusBadRequest, fmt.Sprintf("Error binding body: %s\n", err.Error()))
  }
  if err := loadSongStates(db, updateModel); err != nil {
    return c.String(http.StatusInternalServerError, fmt.Sprintf("Error updating song states: %s\n", err.Error()))
  }
  return c.String(http.StatusOK, "")
}
