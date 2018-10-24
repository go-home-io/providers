package main

import (
	"net/http"
	_ "net/http/pprof"

	"go-home.io/x/server/plugins/api"
)

// PprofAPI is a pprof wrapper.
type PprofAPI struct {
}

// Init makes an attempt to setup a new pprof API extension.
func (*PprofAPI) Init(data *api.InitDataAPI) error {
	data.InternalRootRouter.PathPrefix("/debug/").Handler(http.DefaultServeMux)
	return nil
}

// Routes returns registered routes.
func (*PprofAPI) Routes() []string {
	return []string{"/debug/pprof"}
}

// Unload is responsible for plugin unload.
// This plugin runs on server side only
// and we don't have anything to stop.
func (*PprofAPI) Unload() {
}
