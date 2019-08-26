module go-home.io/x/providers/lock/august

go 1.13

require (
	github.com/pkg/errors v0.8.0
	github.com/vkorn/go-august v0.0.0-20190825051103-aa6e6e31b96f
	go-home.io/x/server/plugins v0.0.0-20190823171444-725318f75f8d
)

replace go-home.io/x/server/plugins => ../../../server/plugins
