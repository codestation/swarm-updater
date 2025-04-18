/*
Copyright 2025 codestation

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
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/docker/cli/cli/connhelper"
	"github.com/docker/docker/client"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/urfave/cli"
)

const (
	DefaultListenAddr = ":8000"
	DefaultTimeout    = 30 * time.Second
)

var blacklist []*regexp.Regexp

// UpdateRequest has a list of images that should be updated on the services that uses them
type UpdateRequest struct {
	Images []string `json:"images"`
}

func run(c *cli.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var opts []client.Opt

	host := c.String("host")
	if host == "" {
		host = os.Getenv("DOCKER_HOST")
	}

	if host != "" {
		if strings.HasPrefix(host, "ssh://") {
			helper, err := connhelper.GetConnectionHelper(host)
			if err != nil {
				return fmt.Errorf("could not connect to SSH host %s: %w", host, err)
			}
			opts = append(opts, client.WithHost(helper.Host))
			opts = append(opts, client.WithDialContext(helper.Dialer))
		} else {
			opts = append(opts, client.WithHost(host))
		}
	}

	opts = append(opts, client.WithTLSClientConfigFromEnv())

	if os.Getenv("DOCKER_API_VERSION") != "" {
		opts = append(opts, client.WithVersionFromEnv())
	} else {
		opts = append(opts, client.WithAPIVersionNegotiation())
	}

	configDir := c.String("config")
	if configDir == "" {
		configDir = os.Getenv("DOCKER_CONFIG")
	}

	swarm, err := NewSwarm(configDir, opts...)
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
		if updateErr := swarm.UpdateServices(ctx); updateErr != nil {
			slog.Error("Failed to update services", "error", updateErr.Error())
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
	e.Use(middleware.KeyAuth(func(key string, _ echo.Context) (bool, error) {
		return key == apiKey, nil
	}))

	e.POST("/apis/swarm/v1/update", func(c echo.Context) error {
		req := &UpdateRequest{}
		if err := c.Bind(req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Bind:"+err.Error())
		}

		if len(req.Images) == 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "No images to update")
		}

		slog.Info("Received update request", "images", strings.Join(req.Images, ","))

		if err := swarm.UpdateServices(c.Request().Context(), req.Images...); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Swarm update:"+err.Error())
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	svr := &http.Server{
		Addr:         c.String("listen"),
		ReadTimeout:  DefaultTimeout,
		WriteTimeout: DefaultTimeout,
	}

	go func() {
		if err := e.StartServer(svr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Failed to start web server", "error", err.Error())
		}
	}()

	cron.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	if err := e.Shutdown(ctx); err != nil {
		slog.Error("Failed to shutdown web server", "error", err.Error())
	}

	cron.Stop()

	return nil
}

func initialize(c *cli.Context) error {
	if c.Bool("label-enable") && (c.IsSet("blacklist") || c.IsSet("blacklist-regex")) {
		slog.Error("Do not define a blacklist if label-enable is enabled")
	}

	slog.Info("Starting Swarm Updater",
		"version", Tag,
		"commit", Revision,
		"date", LastCommit,
		"clean_build", !Modified)

	if c.Bool("debug") {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	if c.IsSet("blacklist") {
		list := c.StringSlice("blacklist")
		for _, entry := range list {
			rule := strings.TrimSpace(entry)
			if rule == "" {
				slog.Warn("Ignoring empty rule in blacklist. Did you leave a trailing comma?y")
				continue
			}

			regex, err := regexp.Compile(rule)
			if err != nil {
				return fmt.Errorf("failed to compile blacklist regex: %w", err)
			}

			blacklist = append(blacklist, regex)
		}

		slog.Debug("Blacklist rules compiled", "count", len(list))
	}

	return nil
}

func printVersion(_ *cli.Context) {
	slog.Info("Starting Swarm Updater",
		"version", Tag,
		"commit", Revision,
		"date", LastCommit,
		"clean_build", !Modified)
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
			Value: "",
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
			Value:  DefaultListenAddr,
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
			Value:  1,
		},
	}

	app.Before = initialize
	app.Action = run

	if err := app.Run(os.Args); err != nil {
		slog.Error("Cannot start program", "error", err.Error())
	}
}
