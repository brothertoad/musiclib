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
		return getArtists(e, c, db)
	})
  e.GET("/albums/:artistId/:state", func(c echo.Context) error {
		return getAlbums(e, c, db)
	})
  e.GET("/songs/:albumId/:state", func(c echo.Context) error {
		return getSongs(e, c, db)
	})
  e.GET("/allsongs/:state", func(c echo.Context) error {
		return getAllSongs(e, c, db)
	})
  e.POST("/updatesongs", func(c echo.Context) error {
		return updateSongStates(e, c, db)
	})

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
  return nil
}

func getArtists(e *echo.Echo, c echo.Context, db *sql.DB) error {
  stateString := c.Param("state")
  state, err := strconv.Atoi(stateString)
  if err != nil {
    e.Logger.Errorf("Can't convert state '%s' to a number\n", stateString)
    return c.String(http.StatusBadRequest, "Can't convert state to a number\n")
  }
  artists, err := loadArtists(db, state)
  if err != nil {
    e.Logger.Errorf("Error loading artists: %s\n", err.Error())
    return c.String(http.StatusBadRequest, "Error loading artists\n")
  }
  return c.JSON(http.StatusOK, artists)
}

func getAlbums(e *echo.Echo, c echo.Context, db *sql.DB) error {
  artistString := c.Param("artistId")
  artistId, err := strconv.Atoi(artistString)
  if err != nil {
    e.Logger.Errorf("Can't convert artistId '%s' to a number\n", artistString)
    return c.String(http.StatusBadRequest, "Can't convert artistId to a number\n")
  }
  stateString := c.Param("state")
  state, err := strconv.Atoi(stateString)
  if err != nil {
    e.Logger.Errorf("Can't convert state '%s' to a number\n", stateString)
    return c.String(http.StatusBadRequest, "Can't convert state to a number\n")
  }
  artists, err := loadAlbums(db, artistId, state)
  if err != nil {
    e.Logger.Errorf("Error loading albums: %s\n", err.Error())
    return c.String(http.StatusBadRequest, "Error loading albums\n")
  }
  return c.JSON(http.StatusOK, artists)
}

func getSongs(e *echo.Echo, c echo.Context, db *sql.DB) error {
  albumString := c.Param("albumId")
  albumId, err := strconv.Atoi(albumString)
  if err != nil {
    e.Logger.Errorf("Can't convert albumId '%s' to a number\n", albumString)
    return c.String(http.StatusBadRequest, "Can't convert albumId to a number\n")
  }
  stateString := c.Param("state")
  state, err := strconv.Atoi(stateString)
  if err != nil {
    e.Logger.Errorf("Can't convert state '%s' to a number\n", stateString)
    return c.String(http.StatusBadRequest, "Can't convert state to a number\n")
  }
  songs, err := loadSongs(db, albumId, state)
  if err != nil {
    e.Logger.Errorf("error loading songs: %s\n", err.Error())
    return c.String(http.StatusBadRequest, "Error loading songs\n")
  }
  return c.JSON(http.StatusOK, songs)
}

func getAllSongs(e *echo.Echo, c echo.Context, db *sql.DB) error {
  stateString := c.Param("state")
  state, err := strconv.Atoi(stateString)
  if err != nil {
    e.Logger.Errorf("Can't convert state '%s' to a number\n", stateString)
    return c.String(http.StatusBadRequest, "Can't convert state to a number\n")
  }
  songs, err := loadAllSongs(db, state)
  if err != nil {
    e.Logger.Errorf("error loading songs: %s\n", err.Error())
    return c.String(http.StatusBadRequest, "Error loading songs\n")
  }
  return c.JSON(http.StatusOK, songs)
}

func updateSongStates(e *echo.Echo, c echo.Context, db *sql.DB) error {
  updateModel := new(UpdateSongStatesModel)
  if err := c.Bind(updateModel); err != nil {
    e.Logger.Errorf("Error binding body: %s\n", err.Error())
    return c.String(http.StatusBadRequest, "Error binding body\n")
  }
  if err := loadSongStates(db, updateModel); err != nil {
    e.Logger.Errorf("Error updating song states: %s\n", err.Error())
    return c.String(http.StatusInternalServerError, "Error updating song states\n")
  }
  return c.String(http.StatusOK, "")
}
