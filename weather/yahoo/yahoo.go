package main

import (
	"errors"
	"strings"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/plugins/helpers"
	"github.com/koron/go-dproxy"
	"github.com/sndnvaps/yahoo_weather_api"
)

// YahooWeather implements IWeather.
type YahooWeather struct {
	Settings *Settings

	state  *device.WeatherState
	logger common.ILoggerProvider
	proxy  dproxy.Proxy
	uom    enums.UOM
}

// Init starts weather device.
func (w *YahooWeather) Init(data *device.InitDataDevice) error {
	w.logger = data.Logger
	w.uom = data.UOM

	w.proxy = yahoo.GetChannelNode(w.Settings.Location)
	if nil == w.proxy {
		err := errors.New("failed to load location")
		w.logger.Error("Failed to load yahoo location", err)
		return err
	}

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
func (w *YahooWeather) Update() (*device.WeatherState, error) {
	w.state = &device.WeatherState{}

	uom := yahoo.GetUnits(w.proxy)
	w.logger.Debug("Pulls atmosphere data from Yahoo weather")
	a := yahoo.GetAtmosphere(w.proxy)

	if enums.SliceContainsProperty(w.Settings.Properties, enums.PropHumidity) {
		w.state.Humidity = a.Humidity
	}

	if enums.SliceContainsProperty(w.Settings.Properties, enums.PropPressure) {
		w.state.Pressure = helpers.UOMConvertString(a.Pressure, enums.PropPressure, uom.Pressure, w.uom)
	}

	if enums.SliceContainsProperty(w.Settings.Properties, enums.PropVisibility) {
		w.state.Visibility = helpers.UOMConvertString(a.Visibility, enums.PropVisibility, uom.Distance, w.uom)
	}

	if enums.SliceContainsProperty(w.Settings.Properties, enums.PropTemperature) {
		w.logger.Debug("Pulls temperature from Yahoo weather")
		c := yahoo.GetConditions(w.proxy)
		w.state.Temperature = helpers.UOMConvertString(c.Temp, enums.PropTemperature, uom.Temperature, w.uom)
	}

	w.logger.Debug("Pulls astronomy data from Yahoo weather")
	as := yahoo.GetAstronomy(w.proxy)
	if enums.SliceContainsProperty(w.Settings.Properties, enums.PropSunset) {
		w.state.Sunset = as.Sunset
	}

	if enums.SliceContainsProperty(w.Settings.Properties, enums.PropSunrise) {
		w.state.Sunrise = as.Sunrise
	}

	w.logger.Debug("Pulls wind data from Yahoo weather")
	wi := yahoo.GetWindInfo(w.proxy)
	if enums.SliceContainsProperty(w.Settings.Properties, enums.PropWindDirection) {
		w.state.WindDirection = wi.Direction
	}

	if enums.SliceContainsProperty(w.Settings.Properties, enums.PropWindSpeed) {
		w.state.WindSpeed = helpers.UOMConvertString(wi.Speed, enums.PropWindSpeed, uom.Speed, w.uom)
	}

	return w.state, nil
}
