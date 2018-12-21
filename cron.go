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
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"github.com/robfig/cron"

	"megpoid.xyz/go/swarm-updater/log"
)

func runCron(schedule string, useLabels bool) error {
	swarm, err := NewSwarm()
	if err != nil {
		return errors.Wrap(err, "cannot instantiate new Docker swarm client")
	}

	swarm.LabelEnable = useLabels
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
				log.Debug("Skipped service update. Already running")
			}

			nextRuns := cronService.Entries()
			if len(nextRuns) > 0 {
				log.Debug("Scheduled next run: " + nextRuns[0].Next.String())
			}
		})

	if err != nil {
		return errors.Wrap(err, "failed to setup cron")
	}

	log.Debug("Configured cron schedule: %s", schedule)

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
