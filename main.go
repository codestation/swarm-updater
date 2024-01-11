/*
Copyright 2018 codestation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/urfave/cli"
	"megpoid.dev/go/swarm-updater/log"
)

var blacklist []*regexp.Regexp

const versionFormatter = `Swarm Updater version: %s, commit: %s, date: %s, clean build: %t`

// UpdateRequest has a list of images that should be updated on the services that uses them
type UpdateRequest struct {
	Images []string `json:"images"`
}

func run(c *cli.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	swarm, err := NewSwarm()
	if err != nil {
		return fmt.Errorf("cannot instantiate new Docker swarm client: %w", err)
	}

	swarm.LabelEnable = c.Bool("label-enable")
	swarm.Blacklist = blacklist
	swarm.MaxThreads = c.Int("max-threads")
	schedule := c.String("schedule")

	// update the services and exit, if requested
	if schedule == "none" {
		return swarm.UpdateServices(ctx)
	}

	cron, err := NewCronService(schedule, func() {
		if err := swarm.UpdateServices(ctx); err != nil {
			log.Printf("Cannot update services: %s", err.Error())
		}
	})

	if err != nil {
		return fmt.Errorf("failed to setup cron, %w", err)
	}

	e := echo.New()
	e.HideBanner = true
	e.Debug = c.Bool("debug")
	e.Use(middleware.Recover())
	apiKey := c.String("apikey")
	e.Use(middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
		return key == apiKey, nil
	}))

	e.POST("/apis/swarm/v1/update", func(c echo.Context) error {
		req := &UpdateRequest{}
		if err := c.Bind(req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Bind:"+err.Error())
		}

		log.Printf("Called update endpoint with images: %s", strings.Join(req.Images, ", "))

		if err := swarm.UpdateServices(c.Request().Context(), req.Images...); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Swarm update:"+err.Error())
		}

		return c.NoContent(http.StatusNoContent)
	})

	svr := &http.Server{
		Addr:         c.String("listen"),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		if err := e.StartServer(svr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Web server failed: %s", err.Error())
		}
	}()

	cron.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	if err := e.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	cron.Stop()

	return nil
}

func initialize(c *cli.Context) error {
	if c.Bool("label-enable") && (c.IsSet("blacklist") || c.IsSet("blacklist-regex")) {
		log.Fatal("Do not define a blacklist if label-enable is enabled")
	}

	log.Printf(versionFormatter, Tag, Revision, LastCommit, Modified)

	err := envConfig(c)
	if err != nil {
		return fmt.Errorf("failed to sync environment: %w", err)
	}

	log.EnableDebug(c.Bool("debug"))

	if c.IsSet("blacklist") {
		list := c.StringSlice("blacklist")
		for _, entry := range list {
			rule := strings.TrimSpace(entry)
			if rule == "" {
				log.Println("Warning: ignoring empty rule in blacklist. Did you leave a trailing comma?")
				continue
			}

			regex, err := regexp.Compile(rule)
			if err != nil {
				return fmt.Errorf("failed to compile blacklist regex: %w", err)
			}

			blacklist = append(blacklist, regex)
		}

		log.Debug("Compiled %d blacklist rules", len(list))
	}

	return nil
}

func printVersion(c *cli.Context) {
	_, _ = fmt.Fprintf(c.App.Writer, versionFormatter, Tag, Revision, LastCommit, Modified)
}

func main() {
	app := cli.NewApp()
	app.Usage = "automatically update Docker services"
	app.Version = Tag
	cli.VersionPrinter = printVersion

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "host, H",
			Usage: "docker host",
			Value: "unix:///var/run/docker.sock",
		},
		cli.BoolFlag{
			Name:  "tlsverify, t",
			Usage: "use TLS and verify the server certificate",
		},
		cli.StringFlag{
			Name:  "config, c",
			Usage: "location of the docker config files",
		},
		cli.StringFlag{
			Name:   "schedule, s",
			Usage:  "cron schedule",
			Value:  "@every 1h",
			EnvVar: "SCHEDULE",
		},
		cli.BoolFlag{
			Name:   "label-enable, l",
			Usage:  fmt.Sprintf("watch services where %s label is set to true", enabledServiceLabel),
			EnvVar: "LABEL_ENABLE",
		},
		cli.StringSliceFlag{
			Name:   "blacklist, b",
			Usage:  "regular expression to match service names to ignore",
			EnvVar: "BLACKLIST",
		},
		cli.BoolFlag{
			Name:   "debug, d",
			Usage:  "enable debug logging",
			EnvVar: "DEBUG",
		},
		cli.StringFlag{
			Name:   "listen, a",
			Usage:  "listen address",
			Value:  ":8000",
			EnvVar: "LISTEN",
		},
		cli.StringFlag{
			Name:   "apikey, k",
			Usage:  "api key to protect endpoint",
			EnvVar: "APIKEY",
		},
		cli.IntFlag{
			Name:   "max-threads, m",
			Usage:  "max threads",
			EnvVar: "MAX_THREADS",
			Value:  5,
		},
	}

	app.Before = initialize
	app.Action = run

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Unrecoverable error: %s", err.Error())
	}
}
