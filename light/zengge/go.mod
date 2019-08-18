module go-home.io/x/providers/light/zengge

require (
	github.com/pkg/errors v0.8.0
	github.com/vikstrous/zengge-lightcontrol v0.0.0-20170104042503-6613d030df9f
	go-home.io/x/server/plugins v0.0.0-20181025030525-18e916b213bc
)

replace go-home.io/x/server/plugins => ../../../server/plugins

go 1.13
