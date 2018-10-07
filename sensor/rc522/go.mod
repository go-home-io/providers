module github.com/go-home-io/providers/sensor/rc522

require (
	github.com/ecc1/gpio v0.0.0-20171107174639-450ac9ea6df7 // indirect
	github.com/ecc1/spi v0.0.0-20180427171038-91ea03fbebdc // indirect
	github.com/go-home-io/server v0.0.0-20181002225757-7899d71e144f
	github.com/jdevelop/golang-rpi-extras v0.0.0-20180927120539-647b09b298b6
	github.com/jdevelop/gpio v0.0.0-20180116031910-0e2cc992019a // indirect
	github.com/sirupsen/logrus v1.0.6
)

replace github.com/go-home-io/server => ../../../server
