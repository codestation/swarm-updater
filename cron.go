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
	"os"
	"os/signal"
	"syscall"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

func runCron(schedule string, useLabels bool) error {
	swarm, err := NewSwarm()
	if err != nil {
		return fmt.Errorf("cannot instantiate new Docker swarm client: %w", err)
	}

	swarm.LabelEnable = useLabels
	swarm.Blacklist = blacklist

	ctx, cancel := context.WithCancel(context.Background())

	tryLockSem := make(chan bool, 1)
	tryLockSem <- true

	// retain non standard seconds support
	cronService := cron.New(cron.WithParser(cron.NewParser(
		cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)))

	_, err = cronService.AddFunc(
		schedule,
		func() {
			select {
			case v := <-tryLockSem:
				defer func() { tryLockSem <- v }()
				if err := swarm.UpdateServices(ctx); err != nil {
					logrus.Printf("Cannot update services: %s", err.Error())
				}
			default:
				logrus.Debug("Skipped service update. Already running")
			}
		})

	if err != nil {
		return fmt.Errorf("failed to setup cron: %w", err)
	}

	logrus.Debugf("Configured cron schedule: %s", schedule)

	cronService.Start()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	<-interrupt

	cancel()

	cronCtx := cronService.Stop()
	<-cronCtx.Done()

	logrus.Println("Waiting for running update to be finished...")
	<-tryLockSem

	return nil
}
