module go-home.io/x/providers/notification/telegram

require (
	github.com/kimrgrey/go-telegram v0.0.0-20170122230828-955a999278a2
	github.com/pkg/errors v0.8.0
	go-home.io/x/server/plugins v0.0.0-20181025030525-18e916b213bc
)

replace go-home.io/x/server/plugins => ../../../server/plugins

go 1.13
