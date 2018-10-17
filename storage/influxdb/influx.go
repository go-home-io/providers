package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/storage"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"
)

// InfluxStorage implements storage provider.
type InfluxStorage struct {
	sync.Mutex
	Settings *Settings

	logger common.ILoggerProvider
	points []*client.Point
}

// Init performs initial test of influxdb connectivity.
//noinspection GoUnhandledErrorResult
func (i *InfluxStorage) Init(data *storage.InitDataStorage) error {
	i.logger = data.Logger
	i.points = make([]*client.Point, 0)

	c, err := i.getClient()
	if err != nil {
		return errors.Wrap(err, "get client failed")
	}

	defer c.Close()
	return nil
}

// Heartbeat stores heartbeat event from a device.
func (i *InfluxStorage) Heartbeat(deviceID string) {
	i.Lock()
	defer i.Unlock()

	i.addPoint(deviceID, map[string]interface{}{"ping": true}, map[string]string{eventToken: heartbeatToken})
}

// State stores state change event from a device.
func (i *InfluxStorage) State(deviceID string, deviceData map[string]interface{}) {
	i.Lock()
	defer i.Unlock()

	i.addPoint(deviceID, deviceData, map[string]string{eventToken: stateToken})
}

// History returns state history records.
func (i *InfluxStorage) History(deviceID string, hrs int) map[string]map[int64]interface{} {
	c, err := i.getClient()
	if err != nil {
		return nil
	}

	q := client.NewQuery(
		fmt.Sprintf( // nolint: gosec
			`SELECT * FROM "%s" WHERE %s='%s' AND time > '%s' - %dh ORDER BY time DESC LIMIT 1000`,
			deviceID,
			eventToken,
			stateToken,
			now().Format(time.RFC3339),
			hrs),
		i.Settings.Database, "ns")
	r, err := c.Query(q)
	if err != nil || r.Error() != nil {
		if err == nil {
			err = r.Error()
		}

		i.logger.Error("Failed to get device history", err, common.LogIDToken, deviceID)
		return nil
	}

	return i.parseInfluxResponse(r, deviceID)
}

// Returns influx client.
func (i *InfluxStorage) getClient() (client.Client, error) {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Username: i.Settings.Username,
		Password: i.Settings.Password,
		Addr:     i.Settings.Address,
	})

	if err != nil {
		i.logger.Error("Failed to instantiate influx client", err)
		return nil, errors.Wrap(err, "http client init failed")
	}

	return c, nil
}

// Adds point to internal cache.
func (i *InfluxStorage) addPoint(deviceID string, deviceData map[string]interface{}, tags map[string]string) {
	point, err := client.NewPoint(deviceID, tags, deviceData, now())

	if err != nil {
		i.logger.Error("Failed to create influx data point", err, common.LogIDToken, deviceID)
		return
	}

	i.points = append(i.points, point)
	i.checkAndSave()
}

// Checks whether amount of cached points is greater than threshold and performs
// save if yes.
//noinspection GoUnhandledErrorResult
func (i *InfluxStorage) checkAndSave() {
	if len(i.points) < i.Settings.BatchSize {
		return
	}

	c, err := i.getClient()
	if err != nil {
		return
	}

	defer c.Close()

	bps, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database: i.Settings.Database,
	})

	if err != nil {
		i.logger.Error("Failed to create influx batch points", err)
		return
	}

	bps.AddPoints(i.points)
	err = c.Write(bps)
	if err != nil {
		i.logger.Error("Failed to store influx data", err)
	} else {
		i.logger.Debug("Finished influx transaction")
	}

	i.points = make([]*client.Point, 0)
}

// Parses influx db query response.
// nolint: gocyclo
func (i *InfluxStorage) parseInfluxResponse(r *client.Response, deviceID string) map[string]map[int64]interface{} {
	result := make(map[string]map[int64]interface{})

	for _, v := range r.Results {
		for _, j := range v.Series {
			columnsLen := len(j.Columns)
			timeColumn := findTimeColumn(j.Columns)

			if -1 == timeColumn {
				i.logger.Warn("Didn't find time column", common.LogIDToken, deviceID)
				continue
			}

			for _, val := range j.Values {
				if timeColumn >= len(val) || nil == val[timeColumn] || columnsLen > len(val) {
					i.logger.Warn("Wrong influx data for device", common.LogIDToken, deviceID)
					continue
				}
				t, err := val[timeColumn].(json.Number).Int64()
				if err != nil {
					i.logger.Warn("Wrong influx time data for device", common.LogIDToken, deviceID)
					continue
				}

				t = t / 1000
				for ii := 0; ii < columnsLen; ii++ {
					if ii == timeColumn || nil == val[ii] {
						continue
					}

					p := j.Columns[ii]

					if eventToken == p || heartbeatToken == p {
						continue
					}

					_, ok := result[p]
					if !ok {
						result[p] = make(map[int64]interface{})
					}

					result[p][t] = val[ii]
				}
			}
		}
	}

	return result
}

// Finds time column in influx db query response.
// If column is not found, returns -1.
func findTimeColumn(columns []string) int {
	columnsLen := len(columns)
	for ii := 0; ii < columnsLen; ii++ {
		if columns[ii] == "time" {
			return ii
		}
	}

	return -1
}
