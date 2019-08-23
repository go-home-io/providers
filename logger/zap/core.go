package main

import (
	"go-home.io/x/server/plugins/common"
)

// IHistoryCore defines zap core with history support.
type IHistoryCore interface {
	Query(*common.LogHistoryRequest) []*common.LogHistoryEntry
}
