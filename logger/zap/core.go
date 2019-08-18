package main

import "go-home.io/x/server/plugins/logger"

// IHistoryCore defines zap core with history support.
type IHistoryCore interface {
	Query(*logger.LogHistoryRequest) []*logger.LogHistoryEntry
}
