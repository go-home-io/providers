module github.com/go-home-io/providers/hub/mqtt

require (
	github.com/eclipse/paho.mqtt.golang v1.1.2-0.20180918140736-ae8614d9932c
	github.com/go-home-io/server v0.0.0-20180813052334-aa78a18bea1b
	github.com/pkg/errors v0.8.0
	golang.org/x/net v0.0.0-20181011144130-49bb7cea24b1 // indirect
)

replace github.com/go-home-io/server => ../../../server

replace golang.org/x/net => golang.org/x/net v0.0.0-20180824045131-faa378e6dbae
