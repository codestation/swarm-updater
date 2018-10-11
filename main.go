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
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/robfig/cron"
	"github.com/urfave/cli"
)

var blacklist []string

var (
	// BuildTime indicates the date when the binary was built (set by -ldflags)
	BuildTime string
	// BuildCommit indicates the git commit of the build
	BuildCommit string
	// AppVersion indicates the application version
	AppVersion = "0.1.0"
)

func run(c *cli.Context) error {
	var schedule string

	if c.IsSet("schedule") {
		schedule = c.String("schedule")
	} else {
		schedule = "@every " + strconv.Itoa(c.Int("interval")) + "s"
	}

	swarm, err := NewSwarm()
	if err != nil {
		return fmt.Errorf("cannot instantiate new Docker swarm client: %s", err.Error())
	}

	swarm.LabelEnable = c.Bool("label-enable")
	swarm.Blacklist = blacklist

	tryLockSem := make(chan bool, 1)
	tryLockSem <- true

	cronService := cron.New()
	err = cronService.AddFunc(
		schedule,
		func() {
			select {
			case v := <-tryLockSem:
				defer func() { tryLockSem <- v }()
				if err := swarm.UpdateServices(); err != nil {
					log.Printf("Cannot update services: %s", err.Error())
				}
			default:
				log.Printf("Skipped service update. Already running")
			}

			nextRuns := cronService.Entries()
			if len(nextRuns) > 0 {
				log.Printf("Scheduled next run: " + nextRuns[0].Next.String())
			}
		})

	if err != nil {
		return fmt.Errorf("failed to setup cron: %s", err.Error())
	}

	cronService.Start()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	signal.Notify(interrupt, syscall.SIGTERM)

	<-interrupt
	cronService.Stop()
	log.Println("Waiting for running update to be finished...")
	<-tryLockSem

	return nil
}

func initialize(c *cli.Context) error {
	if c.IsSet("interval") && c.IsSet("schedule") {
		log.Fatal("Only schedule or interval can be defined, not both")
	}

	if c.Bool("label-enable") && c.IsSet("blacklist") {
		log.Fatal("Do not define a blacklist if label-enable is enabled")
	}

	err := envConfig(c)
	if err != nil {
		return fmt.Errorf("failed to sync environment: %s", err.Error())
	}

	if c.IsSet("blacklist") {
		blacklist = strings.Split(c.String("blacklist"), ",")
		for i := range blacklist {
			blacklist[i] = strings.TrimSpace(blacklist[i])
		}
	}

	log.Printf("Starting swarm-updater %s", AppVersion)

	if len(BuildTime) > 0 {
		log.Printf("Build Time: %s", BuildTime)
		log.Printf("Build Commit: %s", BuildCommit)
	}

	return nil
}

func main() {
	app := cli.NewApp()
	app.Usage = "automatically update Docker services"
	app.Version = AppVersion

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
		cli.IntFlag{
			Name:   "interval, i",
			Value:  300,
			Usage:  "poll interval (in seconds)",
			EnvVar: "INTERVAL",
		},
		cli.StringFlag{
			Name:   "schedule, s",
			Usage:  "cron schedule",
			EnvVar: "SCHEDULE",
		},
		cli.BoolFlag{
			Name:   "label-enable, l",
			Usage:  fmt.Sprintf("watch services where %s label is set to true", serviceLabel),
			EnvVar: "LABEL_ENABLE",
		},
		cli.StringFlag{
			Name:   "blacklist, b",
			Usage:  "comma separated list of services to ignore",
			EnvVar: "BLACKLIST",
		},
	}

	app.Before = initialize
	app.Action = run

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalf("Unrecoverable error: %s", err.Error())
	}
}
