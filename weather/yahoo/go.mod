module go-home.io/x/providers/weather/yahoo

require (
	github.com/koron/go-dproxy v1.2.1
	github.com/pkg/errors v0.8.0
	github.com/sndnvaps/yahoo_weather_api v0.0.0-20181011133646-f11c6dfb2ccf
	go-home.io/x/server v0.0.0-20180813052334-aa78a18bea1b
)

replace go-home.io/x/server => ../../../server
