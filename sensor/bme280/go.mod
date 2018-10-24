module go-home.io/x/providers/sensor/bme280

require (
	github.com/pkg/errors v0.8.0
	go-home.io/x/server v0.0.0-20181002225757-7899d71e144f
	// Later versions have issues with sysfs
	periph.io/x/periph v3.1.1-0.20180811204730-6e2faaa5091f+incompatible
)

replace go-home.io/x/server => ../../../server
