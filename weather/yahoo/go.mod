module go-home.io/x/providers/weather/yahoo

require (
	github.com/vkorn/go-yahoo-weather v0.0.0-20190822023254-5b1021692abc
	go-home.io/x/server/plugins v0.0.0-20181025030525-18e916b213bc
)

replace go-home.io/x/server/plugins => ../../../server/plugins

go 1.13
