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
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/robfig/cron"
	"github.com/urfave/cli"
)

// DockerAPIMinVersion must use at least 1.30 for DistributionInspect support
const DockerAPIMinVersion string = "1.30"

// ServiceLabel is the label to check on docker services to see if it should be updated
const ServiceLabel string = "xyz.megpoid.swarm-updater.enable"

type serviceValidator func(service swarm.Service) bool

var blacklist []string

var (
	// BuildTime indicates the date when the binary was built (set by -ldflags)
	BuildTime string
	// BuildCommit indicates the git commit of the build
	BuildCommit string
	// AppVersion indicates the application version
	AppVersion = "0.1.0"
)

func getServiceValidator(c *cli.Context) serviceValidator {
	if c.Bool("label-enable") {
		return func(service swarm.Service) bool {
			label := service.Spec.Labels[ServiceLabel]
			return strings.ToLower(label) == "true"
		}
	}

	return func(service swarm.Service) bool {
		serviceName := service.Spec.Name
		for _, entry := range blacklist {
			if entry == serviceName {
				return false
			}
		}
		return true
	}
}

func updateServices(validService serviceValidator) error {
	ctx := context.Background()

	dcli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return fmt.Errorf("failed to initialize docker client: %s", err.Error())
	}

	services, err := dcli.ServiceList(ctx, types.ServiceListOptions{})
	if err != nil {
		return fmt.Errorf("failed to get service list: %s", err.Error())
	}

	//discard := ioutil.Discard
	commcli := command.NewDockerCli(os.Stdin, os.Stdout, os.Stderr, false)
	opts := flags.NewClientOptions()
	commcli.Initialize(opts)

	for _, service := range services {
		if validService(service) {
			image := service.Spec.TaskTemplate.ContainerSpec.Image

			// get docker auth
			encodedAuth, err := command.RetrieveAuthTokenFromImage(ctx, commcli, image)

			if err != nil {
				log.Printf("Cannot retrieve auth token from service %s", service.Spec.Name)
			}

			// remove image hash from name
			imageName := strings.Split(image, "@sha")[0]
			service.Spec.TaskTemplate.ContainerSpec.Image = imageName

			response, err := dcli.ServiceUpdate(ctx, service.ID, service.Version,
				service.Spec, types.ServiceUpdateOptions{EncodedRegistryAuth: encodedAuth, QueryRegistry: true})
			if err != nil {
				log.Printf("Cannot update service %s: %s", service.Spec.Name, err.Error())
				continue
			}

			if len(response.Warnings) > 0 {
				for _, warning := range response.Warnings {
					log.Printf("response warning:\n%s", warning)
				}
			}

			updatedService, _, err := dcli.ServiceInspectWithRaw(ctx, service.ID, types.ServiceInspectOptions{})
			if err != nil {
				log.Printf("Cannot inspect service %s to check update status: %s", service.Spec.Name, err.Error())
				continue
			}

			previous := updatedService.PreviousSpec.TaskTemplate.ContainerSpec.Image
			current := updatedService.Spec.TaskTemplate.ContainerSpec.Image

			if previous == current {
				log.Printf("No updates to service %s", service.Spec.Name)
			} else {
				log.Printf("Service %s updated to %s", service.Spec.Name, current)
			}
		}
	}

	return nil
}

func run(c *cli.Context) error {
	validService := getServiceValidator(c)
	var schedule string

	if c.IsSet("schedule") {
		schedule = c.String("schedule")
	} else {
		schedule = "@every " + strconv.Itoa(c.Int("interval")) + "s"
	}

	tryLockSem := make(chan bool, 1)
	tryLockSem <- true

	cronService := cron.New()
	err := cronService.AddFunc(
		schedule,
		func() {
			select {
			case v := <-tryLockSem:
				defer func() { tryLockSem <- v }()
				if err := updateServices(validService); err != nil {
					log.Printf("cannot update services: %s", err.Error())
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
		return fmt.Errorf("falied to setup cron: %s", err.Error())
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
			Name:   "host, H",
			Usage:  "docker host",
			Value:  "unix:///var/run/docker.sock",
		},
		cli.BoolFlag{
			Name:   "tlsverify, t",
			Usage:  "use TLS and verify the server certificate",
		},
		cli.StringFlag{
			Name:   "config, c",
			Usage:  "location of the docker config files",
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
			Usage:  fmt.Sprintf("watch services where %s label is set to true", ServiceLabel),
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
