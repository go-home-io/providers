module go-home.io/x/providers/camera/web

require (
	github.com/bdlm/errors v0.1.1
	github.com/bdlm/log v0.1.10
	github.com/bdlm/std v0.0.0-20180829231636-9f706d297d7e // indirect
	github.com/gorilla/websocket v1.3.0 // indirect
	github.com/mkenney/go-chrome v1.0.0-rc6
	github.com/pkg/errors v0.8.0
	go-home.io/x/server/plugins v0.0.0-20181025030525-18e916b213bc
	golang.org/x/crypto v0.0.0-20180904163835-0709b304e793 // indirect
	golang.org/x/sys v0.0.0-20180905080454-ebe1bf3edb33 // indirect
)

replace go-home.io/x/server/plugins => ../../../server/plugins
