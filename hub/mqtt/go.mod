module go-home.io/x/providers/hub/mqtt

require (
	github.com/eclipse/paho.mqtt.golang v1.1.2-0.20180918140736-ae8614d9932c
	github.com/pkg/errors v0.8.0
	go-home.io/x/server/plugins v0.0.0-20181025030525-18e916b213bc
	golang.org/x/net v0.0.0-20181011144130-49bb7cea24b1 // indirect
)

replace go-home.io/x/server/plugins => ../../../server/plugins

replace golang.org/x/net => golang.org/x/net v0.0.0-20180824045131-faa378e6dbae

go 1.13
