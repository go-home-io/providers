module github.com/go-home-io/providers/sensor/rc522

require (
	github.com/ecc1/gpio v0.0.0-20171107174639-450ac9ea6df7 // indirect
	github.com/ecc1/spi v0.0.0-20180427171038-91ea03fbebdc // indirect
	github.com/go-home-io/server v0.0.0-20181002225757-7899d71e144f
	github.com/jdevelop/golang-rpi-extras v0.0.0-20181010003844-c75e8edb0d6f
	github.com/jdevelop/gpio v0.0.0-20180116031910-0e2cc992019a // indirect
	github.com/pkg/errors v0.8.0
	github.com/sirupsen/logrus v1.0.6
	golang.org/x/sys v0.0.0-20180909124046-d0be0721c37e // indirect
)

replace github.com/go-home-io/server => ../../../server

replace golang.org/x/sys => golang.org/x/sys v0.0.0-20180821140842-3b58ed4ad339

replace github.com/sirupsen/logrus => github.com/sirupsen/logrus v1.1.1

replace golang.org/x/crypto => golang.org/x/crypto v0.0.0-20180820150726-614d502a4dac
