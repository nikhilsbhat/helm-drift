package pkg

import (
	"strings"

	log "github.com/sirupsen/logrus"
)

func (drift *Drift) SetLogger(logLevel string) {
	logger := log.New()
	logger.SetLevel(GetLoglevel(logLevel))
	logger.WithField("helm-drifts", true)
	logger.SetFormatter(&log.JSONFormatter{})
	drift.log = logger
}

// GetLoglevel sets the loglevel to the kind of log asked for.
func GetLoglevel(level string) log.Level {
	switch strings.ToLower(level) {
	case log.WarnLevel.String():
		return log.WarnLevel
	case log.DebugLevel.String():
		return log.DebugLevel
	case log.TraceLevel.String():
		return log.TraceLevel
	case log.FatalLevel.String():
		return log.FatalLevel
	case log.ErrorLevel.String():
		return log.ErrorLevel
	default:
		return log.InfoLevel
	}
}
