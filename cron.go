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
	"log/slog"

	"github.com/robfig/cron/v3"
)

// CronService holds the instantiated cron service.
type CronService struct {
	cronService *cron.Cron
	tryLockSem  chan bool
}

// NewCronService creates a new cron for the specified function.
func NewCronService(schedule string, cronFunc func()) (*CronService, error) {
	cronService := cron.New(cron.WithParser(cron.NewParser(
		cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)))

	tryLockSem := make(chan bool, 1)
	tryLockSem <- true

	_, err := cronService.AddFunc(
		schedule,
		func() {
			select {
			case v := <-tryLockSem:
				defer func() { tryLockSem <- v }()
				cronFunc()
			default:
				slog.Debug("Skipping cron schedule. Already running")
			}
		})
	if err != nil {
		return nil, err
	}

	slog.Debug("Configured cron schedule", "schedule", schedule)

	return &CronService{
		cronService: cronService,
		tryLockSem:  tryLockSem,
	}, nil
}

// Start initiates the cron schedule.
func (c *CronService) Start() {
	c.cronService.Start()
}

// Stop cancel the schedule and wait until the currently running function is finished.
func (c *CronService) Stop() {
	ctx := c.cronService.Stop()
	<-ctx.Done()

	slog.Info("Waiting for running update to be finished...")
	<-c.tryLockSem
}
