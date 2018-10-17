package main

import (
	"time"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/trigger"
)

// CronTrigger implements cron trigger interface.
type CronTrigger struct {
	Settings *Settings

	logger    common.ILoggerProvider
	triggered chan interface{}
	next      time.Time
}

// Init starts internal cycle of cron events.
func (t *CronTrigger) Init(data *trigger.InitDataTrigger) error {
	t.logger = data.Logger
	t.triggered = data.Triggered

	t.next = t.Settings.expr.Next(time.Now())
	go t.wait()

	return nil
}

// Waits till the next cycle.
func (t *CronTrigger) wait() {
	for {
		t.logger.Info("Calculated next time", common.LogTimeToken, t.next.Format(time.Stamp))
		time.Sleep(time.Until(t.next))
		t.logger.Debug("Triggering due to time", common.LogTimeToken, t.next.Format(time.Stamp))
		t.triggered <- true
		t.next = t.Settings.expr.Next(time.Now())
		// Just in case to prevent overlaps.
		time.Sleep(1 * time.Second)
	}
}
