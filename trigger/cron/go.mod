module go-home.io/x/providers/trigger/cron

require (
	github.com/gorhill/cronexpr v0.0.0-20180427100037-88b0669f7d75
	go-home.io/x/server/plugins v0.0.0-20181025030525-18e916b213bc
)

replace go-home.io/x/server/plugins => ../../../server/plugins

go 1.13
