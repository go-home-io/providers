module go-home.io/x/providers/hub/hue

require (
	github.com/amimof/huego v0.0.0-20190527081818-99babfa78560
	github.com/pkg/errors v0.8.0
	go-home.io/x/server/plugins v0.0.0-20181025030525-18e916b213bc
)

replace go-home.io/x/server/plugins => ../../../server/plugins

go 1.13
