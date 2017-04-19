package common

import (
	"github.com/op/go-logging"
	"os"
)

// SetupLogging - verbosity 0=error, 1=info, 2=debug
func SetupLogging(verbosity int) {
	errBackend := logging.NewLogBackend(os.Stderr, "", 0)
	errFormat := logging.MustStringFormatter(
		`%{color}%{shortfunc} ▶ %{level:.5s} %{color:reset} %{message}`,
	)
	errFormatter := logging.NewBackendFormatter(errBackend, errFormat)
	errLeveled := logging.AddModuleLevel(errFormatter)
	errLeveled.SetLevel(logging.ERROR, "")

	warnFormat := logging.MustStringFormatter(
		`%{color}%{message}%{color:reset}`,
	)
	warnFormatter := logging.NewBackendFormatter(errBackend, warnFormat)
	warnLeveled := logging.AddModuleLevel(notToExceedLevel(logging.WARNING, warnFormatter))
	warnLeveled.SetLevel(logging.WARNING, "")

	if verbosity >= 1 {
		infoBackend := logging.NewLogBackend(os.Stdout, "", 0)
		infoFormat := logging.MustStringFormatter(
			`%{color}%{message}%{color:reset}`,
		)
		infoFormatter := logging.NewBackendFormatter(infoBackend, infoFormat)
		infoLeveled := logging.AddModuleLevel(notToExceedLevel(logging.INFO, infoFormatter))
		infoLeveled.SetLevel(logging.INFO, "")

		if verbosity >= 2 {
			debugBackend := logging.NewLogBackend(os.Stdout, "", 0)
			debugFormat := logging.MustStringFormatter(
				`%{color}%{time:15:04:05.000} %{module} ▶ %{id:03x}%{color:reset} %{message}`,
			)
			debugFormatter := logging.NewBackendFormatter(debugBackend, debugFormat)
			debugLeveled := logging.AddModuleLevel(notToExceedLevel(logging.DEBUG, debugFormatter))
			debugLeveled.SetLevel(logging.DEBUG, "")

			logging.SetBackend(debugLeveled, infoLeveled, warnLeveled, errLeveled)
		} else {
			logging.SetBackend(infoLeveled, warnLeveled, errLeveled)
		}
	} else {
		logging.SetBackend(warnLeveled, errLeveled)
	}

}

func notToExceedLevel(notToExceed logging.Level, delegate logging.Backend) logging.Backend {
	return &notToExceedLevelBackend{
		delegate:         delegate,
		notToExceedLevel: notToExceed,
	}
}

type notToExceedLevelBackend struct {
	delegate         logging.Backend
	notToExceedLevel logging.Level
}

func (b *notToExceedLevelBackend) Log(l logging.Level, i int, r *logging.Record) error {
	if l < b.notToExceedLevel {
		return nil
	}
	return b.delegate.Log(l, i, r)
}
