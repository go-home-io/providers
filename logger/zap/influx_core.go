package main

import (
	"fmt"
	"sync"

	"github.com/influxdata/influxdb/client/v2"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/logger"
	"go.uber.org/zap/zapcore"
)

const (
	// Influx key name.
	pointKey = "zap"
)

type influxCore struct {
	sync.Mutex

	zapcore.LevelEnabler
	settings *InfluxSettings

	points []*client.Point
}

func newInfluxCore(enabler zapcore.LevelEnabler, settings *InfluxSettings) *influxCore {
	return &influxCore{
		LevelEnabler: enabler,
		settings:     settings,
	}
}

func (i *influxCore) Query(*logger.LogHistoryRequest) []*logger.LogHistoryEntry {
	panic("implement me")
}

// With is not used.
func (i *influxCore) With([]zapcore.Field) zapcore.Core {
	return i
}

func (i *influxCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if i.Enabled(entry.Level) {
		return checked.AddCore(entry, i)
	}
	return checked
}

func (i *influxCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	f := make(map[string]interface{})
	t := make(map[string]string)

	for _, v := range fields {
		if v.Key == common.LogSystemToken || v.Key == common.LogProviderToken {
			t[v.Key] = v.String
			continue
		}

		// We using strings only.
		f[v.Key] = v.String
	}

	p, err := client.NewPoint(pointKey, t, f, entry.Time)
	if err != nil {
		fmt.Printf("Failed to create InfluxDB point for ZAP: %s", err.Error())
		return err
	}

	i.Lock()
	i.points = append(i.points, p)
	i.Unlock()

	i.checkAndSave()
	return nil
}

func (i *influxCore) Sync() error {
	i.performSave()
	return nil
}

// Returns influx client.
func (i *influxCore) getClient() client.Client {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Username: i.settings.Username,
		Password: i.settings.Password,
		Addr:     i.settings.Address,
	})

	if err != nil {
		return nil
	}

	return c
}

// Checks whether amount of cached points is greater than threshold and performs
// save if yes.
func (i *influxCore) checkAndSave() {
	i.Lock()
	defer i.Unlock()

	if len(i.points) < i.settings.BatchSize {
		return
	}

	go i.performSave()
}

// Performs actual save
//noinspection GoUnhandledErrorResult
func (i *influxCore) performSave() {
	i.Lock()
	defer i.Unlock()

	c := i.getClient()
	if nil == c {
		return
	}

	defer c.Close() // nolint: errcheck

	bps, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database: i.settings.Database,
	})

	if err != nil {
		return
	}

	bps.AddPoints(i.points)
	err = c.Write(bps)
	if err != nil {
		return
	}

	i.points = make([]*client.Point, 0)
}
