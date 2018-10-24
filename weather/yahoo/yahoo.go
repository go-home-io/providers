package main

import (
	"strings"

	"github.com/koron/go-dproxy"
	"github.com/pkg/errors"
	"github.com/sndnvaps/yahoo_weather_api"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/device"
	"go-home.io/x/server/plugins/device/enums"
	"go-home.io/x/server/plugins/helpers"
)

// YahooWeather implements IWeather.
type YahooWeather struct {
	Settings *Settings

	state  *device.WeatherState
	logger common.ILoggerProvider
	uom    enums.UOM
}

// Init starts weather device.
func (w *YahooWeather) Init(data *device.InitDataDevice) error {
	w.logger = data.Logger
	w.uom = data.UOM

	_, err := w.getProxy()
	if err != nil {
		return errors.Wrap(err, "get proxy failed")
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
func (w *YahooWeather) Update() (st *device.WeatherState, e error) {
	proxy, err := w.getProxy()
	if err != nil {
		return nil, errors.Wrap(err, "get proxy failed")
	}

	w.state = &device.WeatherState{}

	defer func() {
		if r := recover(); r != nil {
			st = nil
			e = errors.New("failed to call Yahoo API")
			w.logger.Error("Yahoo weather not responding", e)
		}
	}()

	uom := yahoo.GetUnits(proxy)
	w.logger.Debug("Pulls atmosphere data from Yahoo weather")
	a := yahoo.GetAtmosphere(proxy)

	if enums.SliceContainsProperty(w.Settings.Properties, enums.PropHumidity) {
		w.state.Humidity = a.Humidity
	}

	if enums.SliceContainsProperty(w.Settings.Properties, enums.PropPressure) {
		// Yahoo always reports pressure in mb.
		w.state.Pressure = helpers.UOMConvert(a.Pressure, enums.PropPressure, enums.UOMMetric, w.uom)
	}

	if enums.SliceContainsProperty(w.Settings.Properties, enums.PropVisibility) {
		w.state.Visibility = helpers.UOMConvertString(a.Visibility, enums.PropVisibility, uom.Distance, w.uom)
	}

	if enums.SliceContainsProperty(w.Settings.Properties, enums.PropTemperature) {
		w.logger.Debug("Pulls temperature from Yahoo weather")
		c := yahoo.GetConditions(proxy)
		w.state.Temperature = helpers.UOMConvertString(c.Temp, enums.PropTemperature, uom.Temperature, w.uom)
	}

	w.logger.Debug("Pulls astronomy data from Yahoo weather")
	as := yahoo.GetAstronomy(proxy)
	if enums.SliceContainsProperty(w.Settings.Properties, enums.PropSunset) {
		w.state.Sunset = as.Sunset
	}

	if enums.SliceContainsProperty(w.Settings.Properties, enums.PropSunrise) {
		w.state.Sunrise = as.Sunrise
	}

	w.logger.Debug("Pulls wind data from Yahoo weather")
	wi := yahoo.GetWindInfo(proxy)
	if enums.SliceContainsProperty(w.Settings.Properties, enums.PropWindDirection) {
		w.state.WindDirection = wi.Direction
	}

	if enums.SliceContainsProperty(w.Settings.Properties, enums.PropWindSpeed) {
		w.state.WindSpeed = helpers.UOMConvertString(wi.Speed, enums.PropWindSpeed, uom.Speed, w.uom)
	}

	return w.state, nil
}

// Returns object proxy.
func (w *YahooWeather) getProxy() (dproxy.Proxy, error) {
	proxy := yahoo.GetChannelNode(w.Settings.Location)
	if nil == proxy {
		err := errors.New("failed to load location")
		w.logger.Error("Failed to load yahoo location", err)
		return nil, err
	}

	return proxy, nil
}
