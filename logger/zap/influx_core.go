package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/helpers"
	"go.uber.org/zap/zapcore"
)

const (
	// Influx key name.
	pointKey = "zap"
	// Log level key name.
	logLevelKey = "lvl"
	// Log message key name.
	logMessageKey = "msg"
)

var (
	// Known influx keys.
	knownKeys = []string{common.LogSystemToken, common.LogProviderToken, common.LogIDToken, common.LogWorkerToken,
		logMessageKey, logLevelKey, "time"}
)

// Zap Core implementation.
type influxCore struct {
	sync.Mutex

	zapcore.LevelEnabler
	settings *InfluxSettings

	points []*client.Point
}

// Constructs a new core.
func newInfluxCore(enabler zapcore.LevelEnabler, settings *InfluxSettings) *influxCore {
	return &influxCore{
		LevelEnabler: enabler,
		settings:     settings,
	}
}

// CreateDatabase insures that database exists.
func (i *influxCore) CreateDatabase() error {
	c := i.getClient()
	if nil == c {
		return errors.New("failed to provision influxDB client")
	}

	defer c.Close() // nolint: errcheck

	r, err := c.Query(client.NewQuery(
		fmt.Sprintf("CREATE DATABASE %s WITH DURATION %s",
			i.settings.Database, i.settings.Retention), i.settings.Database, "ns"))

	if err == nil && r.Error() != nil {
		err = r.Error()
	}

	return err
}

// With is not used.
func (i *influxCore) With([]zapcore.Field) zapcore.Core {
	return i
}

// Check verifies whether core is enabled.
func (i *influxCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if i.Enabled(entry.Level) {
		return checked.AddCore(entry, i)
	}
	return checked
}

// Write writes log entry to the database.
func (i *influxCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	f := make(map[string]interface{})
	t := make(map[string]string)

	for _, v := range fields {
		if v.Key == common.LogSystemToken || v.Key == common.LogProviderToken ||
			v.Key == common.LogIDToken || v.Key == common.LogWorkerToken {
			t[v.Key] = v.String
			continue
		}

		// We using strings only.
		f[v.Key] = v.String
	}

	t[logLevelKey] = entry.Level.String()
	f[logMessageKey] = entry.Message

	p, err := client.NewPoint(pointKey, t, f, entry.Time.UTC())
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

// Sync performs sync.
func (i *influxCore) Sync() error {
	i.performSave()
	return nil
}

// Query makes database query.
func (i *influxCore) Query(r *common.LogHistoryRequest) []*common.LogHistoryEntry {
	res := make([]*common.LogHistoryEntry, 0)

	c := i.getClient()
	if nil == c {
		return res
	}

	rString := fmt.Sprintf(`SELECT * FROM "%s" WHERE time < '%s'`, // nolint: gosec
		pointKey, time.Unix(r.ToUTC, 0).Format(time.RFC3339))

	if r.FromUTC > 0 {
		rString += fmt.Sprintf(` AND time > '%s'`, time.Unix(r.FromUTC, 0).Format(time.RFC3339))
	}

	if "" != r.LogLevel {
		rString += fmt.Sprintf(` AND %s='%s'`, logLevelKey, r.LogLevel)
	}

	if "" != r.DeviceID {
		rString += fmt.Sprintf(` AND %s='%s'`, common.LogIDToken, r.DeviceID)
	}

	if "" != r.Provider {
		rString += fmt.Sprintf(` AND %s='%s'`, common.LogProviderToken, r.Provider)
	}

	if "" != r.System {
		rString += fmt.Sprintf(` AND %s='%s'`, common.LogSystemToken, r.System)
	}

	if "" != r.WorkerID {
		rString += fmt.Sprintf(` AND %s='%s'`, common.LogWorkerToken, r.WorkerID)
	}

	rString += " ORDER BY time DESC LIMIT 1000"

	defer c.Close() // nolint: errcheck

	q := client.NewQuery(rString, i.settings.Database, "ns")
	qr, err := c.Query(q)

	if err != nil || qr.Error() != nil {
		return res
	}

	res = i.parseInfluxResponse(qr)
	return res
}

// Parsing actual influx data
func (i *influxCore) parseInfluxResponse(r *client.Response) []*common.LogHistoryEntry {
	res := make([]*common.LogHistoryEntry, 0)
	for _, v := range r.Results {
		for _, j := range v.Series {
			known, others := parseInfluxColumns(j.Columns)

			for _, val := range j.Values {
				timestamp, err := val[known["time"]].(json.Number).Int64()
				if err != nil {
					continue
				}
				e := &common.LogHistoryEntry{
					UTCTimestamp: timestamp,
					LogLevel:     interfaceToString(val[known[logLevelKey]]),
					System:       interfaceToString(val[known[common.LogSystemToken]]),
					DeviceID:     interfaceToString(val[known[common.LogIDToken]]),
					Provider:     interfaceToString(val[known[common.LogProviderToken]]),
					WorkerID:     interfaceToString(val[known[common.LogWorkerToken]]),
					Message:      interfaceToString(val[known[logMessageKey]]),
					Properties:   make(map[string]string),
				}

				for colN, colI := range others {
					if nil == val[colI] {
						continue
					}

					e.Properties[colN] = val[colI].(string)
				}

				res = append(res, e)
			}
		}
	}

	return res
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

// Converting interface{} to string.
func interfaceToString(i interface{}) string {
	if nil == i {
		return ""
	}

	t, ok := i.(string)

	if !ok {
		return ""
	}

	return t
}

// Processing series columns.
func parseInfluxColumns(columns []string) (map[string]int, map[string]int) {
	knownColumns := make(map[string]int)
	otherColumns := make(map[string]int)
	for k, v := range columns {
		if helpers.SliceContainsString(knownKeys, v) {
			knownColumns[v] = k
		} else {
			otherColumns[v] = k
		}
	}

	return knownColumns, otherColumns
}
