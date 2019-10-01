module go-home.io/x/providers/hub/hue

require (
	github.com/amimof/huego v0.0.0-20190527081818-99babfa78560
	github.com/pkg/errors v0.8.0
	go-home.io/x/server/plugins v0.0.0-20181025030525-18e916b213bc
	gopkg.in/jarcoal/httpmock.v1 v1.0.0-00010101000000-000000000000 // indirect
)

replace go-home.io/x/server/plugins => ../../../server/plugins

replace gopkg.in/jarcoal/httpmock.v1 => github.com/jarcoal/httpmock v1.0.4

go 1.13
