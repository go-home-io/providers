module go-home.io/x/providers/hub/mqtt

require (
	github.com/eclipse/paho.mqtt.golang v1.1.2-0.20180918140736-ae8614d9932c
	github.com/pkg/errors v0.8.0
	go-home.io/x/server v0.0.0-20180813052334-aa78a18bea1b
	golang.org/x/net v0.0.0-20181011144130-49bb7cea24b1 // indirect
)

replace go-home.io/x/server => ../../../server

replace golang.org/x/net => golang.org/x/net v0.0.0-20180824045131-faa378e6dbae
