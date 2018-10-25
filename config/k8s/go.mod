module go-home.io/x/providers/config/k8s

require (
	github.com/ericchiang/k8s v1.1.0
	github.com/golang/protobuf v1.2.0 // indirect
	github.com/pkg/errors v0.8.0
	go-home.io/x/server/plugins v0.0.0-20181025030525-18e916b213bc
	golang.org/x/net v0.0.0-20181011144130-49bb7cea24b1 // indirect
	golang.org/x/sync v0.0.0-20180314180146-1d60e4601c6f // indirect
	golang.org/x/text v0.3.0 // indirect
)

replace go-home.io/x/server/plugins => ../../../server/plugins

replace golang.org/x/net => golang.org/x/net v0.0.0-20180824045131-faa378e6dbae
