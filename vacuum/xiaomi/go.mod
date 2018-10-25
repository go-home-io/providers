module go-home.io/x/providers/vacuum/xiaomi

require (
	github.com/benbjohnson/clock v0.0.0-20161215174838-7dc76406b6d3 // indirect
	github.com/lunixbochs/struc v0.0.0-20180408203800-02e4c2afbb2a // indirect
	github.com/nickw444/miio-go v0.0.0-20180926063007-ce79bf638b2e // indirect
	github.com/pkg/errors v0.8.0
	github.com/sirupsen/logrus v1.1.1 // indirect
	github.com/vkorn/go-miio v0.0.0-20180929223642-adf1adb6425f
	go-home.io/x/server/plugins v0.0.0-20181025003827-3ceb9900099c
)

replace go-home.io/x/server/plugins => ../../../server/plugins
