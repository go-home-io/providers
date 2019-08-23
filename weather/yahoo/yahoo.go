package main

import (
	"strings"

	yahoo "github.com/vkorn/go-yahoo-weather"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/device"
	"go-home.io/x/server/plugins/device/enums"
	"go-home.io/x/server/plugins/helpers"
)

// YahooWeather implements IWeather.
type YahooWeather struct {
	Settings *Settings

	provider *yahoo.YahooWeatherProvider
	state    *device.WeatherState
	logger   common.ILoggerProvider
	uom      enums.UOM
}

// Init starts weather device.
func (w *YahooWeather) Init(data *device.InitDataDevice) error {
	w.logger = data.Logger
	w.uom = data.UOM
	w.provider = yahoo.NewProvider(w.Settings.AppID, w.Settings.ClientID, w.Settings.ClientSecret)
	yahoo.MinUpdateTimeoutSeconds = int64(w.Settings.PollingInterval*60 - 30)

	return nil
}

// Unload is not used.
func (w *YahooWeather) Unload() {
}

// GetName returns first part of the location.
func (w *YahooWeather) GetName() string {
	parts := strings.Split(w.Settings.Location, ",")
	if 0 == len(parts) {
		return "yahoo"
	}

	return parts[0]
}

// GetSpec returns weather spec.
func (w *YahooWeather) GetSpec() *device.Spec {
	return &device.Spec{
		SupportedCommands:   []enums.Command{},
		SupportedProperties: w.Settings.Properties,
		UpdatePeriod:        w.Settings.updateInterval,
	}
}

// Load performs first update.
func (w *YahooWeather) Load() (*device.WeatherState, error) {
	return w.Update()
}

// Update pulls updates from yahoo weather.
func (w *YahooWeather) Update() (st *device.WeatherState, e error) {

	w.logger.Debug("Pulling weather data from Yahoo")

	data, err := w.provider.Query(w.Settings.Location, yahoo.Imperial)
	if err != nil {
		w.logger.Error("Failed to query Yahoo weather", err)
		return nil, err
	}

	w.state = &device.WeatherState{
		Humidity: helpers.UOMConvert(float64(data.Observation.Atmosphere.Humidity),
			enums.PropHumidity, enums.UOMImperial, w.uom),
		Pressure: helpers.UOMConvert(float64(data.Observation.Atmosphere.Pressure),
			enums.PropPressure, enums.UOMImperial, w.uom),
		Visibility: helpers.UOMConvert(float64(data.Observation.Atmosphere.Visibility),
			enums.PropVisibility, enums.UOMImperial, w.uom),
		WindDirection: float64(data.Observation.Wind.Direction),
		WindSpeed:     float64(data.Observation.Wind.Speed),
		Temperature: helpers.UOMConvert(float64(data.Observation.Condition.Temperature),
			enums.PropTemperature, enums.UOMImperial, w.uom),
		Sunrise:     data.Observation.Astronomy.Sunrise,
		Sunset:      data.Observation.Astronomy.Sunset,
		Description: data.Observation.Condition.Text,
	}

	return w.state, nil
}
