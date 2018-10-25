module go-home.io/x/providers/bus/nsq

require (
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db // indirect
	github.com/nsqio/go-nsq v1.0.7
	github.com/pkg/errors v0.8.0
	go-home.io/x/server/plugins v0.0.0-20181025003827-3ceb9900099c
)

replace go-home.io/x/server/plugins => ../../../server/plugins
