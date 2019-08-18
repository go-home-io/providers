module go-home.io/x/providers/logger/zap

require (
	github.com/influxdata/influxdb v1.6.1
	github.com/pkg/errors v0.8.0
	go-home.io/x/server/plugins v0.0.0-20181025030525-18e916b213bc
	go.uber.org/atomic v1.3.2 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.8.0
)

replace go-home.io/x/server/plugins => ../../../server/plugins

go 1.13
