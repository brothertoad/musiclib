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
		return getArtistsForREST(c, db)
	})

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
  return nil
}

func getArtistsForREST(c echo.Context, db *sql.DB) error {
  stateString := c.Param("state")
  state, err := strconv.Atoi(stateString)
  if err != nil {
    c.String(http.StatusBadRequest, fmt.Sprintf("Can't convert state '%s' to a number\n", stateString))
    return nil
  }
  c.String(http.StatusOK, fmt.Sprintf("state is %d\n", state))
  return nil
}
