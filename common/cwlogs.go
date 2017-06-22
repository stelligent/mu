package common

import (
	"time"
)

// LogsViewer for viewing cloudwatch logs
type LogsViewer interface {
	ViewLogs(logGroup string, searchDuration time.Duration, follow bool, filter string, callback func(string, string, int64)) error
}

// LogsManager composite of all logs capabilities
type LogsManager interface {
	LogsViewer
}
