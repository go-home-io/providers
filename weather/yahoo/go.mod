module go-home.io/x/providers/weather/yahoo

require (
	bou.ke/monkey v1.0.1 // indirect
	github.com/koron/go-dproxy v1.2.1
	github.com/pkg/errors v0.8.0
	github.com/sndnvaps/yahoo_weather_api v0.0.0-20181011133646-f11c6dfb2ccf
	go-home.io/x/server/plugins v0.0.0-20181025003827-3ceb9900099c
)

replace go-home.io/x/server/plugins => ../../../server/plugins
