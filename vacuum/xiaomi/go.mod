module github.com/go-home-io/providers/vacuum/xiaomi

require (
	github.com/benbjohnson/clock v0.0.0-20161215174838-7dc76406b6d3 // indirect
	github.com/go-home-io/server v0.0.0-20180826010307-f458aa49cdc6
	github.com/lunixbochs/struc v0.0.0-20180408203800-02e4c2afbb2a // indirect
	github.com/nickw444/miio-go v0.0.0-20180729083032-52b65c4114f1 // indirect
	github.com/pkg/errors v0.8.0
	github.com/sirupsen/logrus v1.1.1 // indirect
	github.com/vkorn/go-miio v0.0.0-20180929223642-adf1adb6425f
)

replace github.com/go-home-io/server => ../../../server
