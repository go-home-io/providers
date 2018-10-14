module github.com/go-home-io/providers/sensor/rc522

require (
	github.com/ecc1/gpio v0.0.0-20171107174639-450ac9ea6df7 // indirect
	github.com/ecc1/spi v0.0.0-20180427171038-91ea03fbebdc // indirect
	github.com/go-home-io/server v0.0.0-20181002225757-7899d71e144f
	github.com/jdevelop/golang-rpi-extras v0.0.0-20181010003844-c75e8edb0d6f
	github.com/jdevelop/gpio v0.0.0-20180116031910-0e2cc992019a // indirect
	github.com/onsi/gomega v1.4.2 // indirect
	github.com/pkg/errors v0.8.0
	github.com/sirupsen/logrus v1.0.6
	gopkg.in/airbrake/gobrake.v2 v2.0.9 // indirect
	gopkg.in/gemnasium/logrus-airbrake-hook.v2 v2.1.2 // indirect
)

replace github.com/go-home-io/server => ../../../server
