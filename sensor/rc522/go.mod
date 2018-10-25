module go-home.io/x/providers/sensor/rc522

require (
	github.com/pkg/errors v0.8.0
	go-home.io/x/server/plugins v0.0.0-20181025003827-3ceb9900099c
	// Later versions have issues with sysfs
	periph.io/x/periph v3.1.1-0.20180811204730-6e2faaa5091f+incompatible
)

replace go-home.io/x/server/plugins => ../../../server/plugins
