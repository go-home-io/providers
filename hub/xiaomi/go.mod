module github.com/go-home-io/providers/hub/xiaomi

require (
	github.com/benbjohnson/clock v0.0.0-20161215174838-7dc76406b6d3 // indirect
	github.com/go-home-io/server v0.0.0-20180813052334-aa78a18bea1b
	github.com/lunixbochs/struc v0.0.0-20180408203800-02e4c2afbb2a // indirect
	github.com/nickw444/miio-go v0.0.0-20180729083032-52b65c4114f1 // indirect
	github.com/pkg/errors v0.8.0
	github.com/sirupsen/logrus v1.1.1 // indirect
	github.com/vkorn/go-miio v0.0.0-20180929223642-adf1adb6425f
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

replace github.com/go-home-io/server => ../../../server

replace golang.org/x/sys => golang.org/x/sys v0.0.0-20180821140842-3b58ed4ad339

replace golang.org/x/crypto => golang.org/x/crypto v0.0.0-20180820150726-614d502a4dac

replace gopkg.in/check.v1 => gopkg.in/check.v1 v0.0.0-20161208181325-20d25e280405
