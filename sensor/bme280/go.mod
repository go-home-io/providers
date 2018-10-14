module github.com/go-home-io/providers/sensor/bme280

require (
	github.com/go-home-io/server v0.0.0-20181002225757-7899d71e144f
	github.com/maciej/bme280 v0.0.0-20180217140430-f1e3fb012361
	github.com/pkg/errors v0.8.0
	golang.org/x/exp v0.0.0-20180907224206-e88728d35e99
)

replace github.com/go-home-io/server => ../../../server
