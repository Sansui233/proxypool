package log

import (
	log "github.com/sirupsen/logrus"
)

type LogLevel int

const (
	TRACE LogLevel = iota
	DEBUG
	INFO
	WARNING
	ERROR
)

var (
	levelMapping = map[LogLevel]log.Level{
		TRACE:   log.TraceLevel,
		DEBUG:   log.DebugLevel,
		INFO:    log.InfoLevel,
		WARNING: log.WarnLevel,
		ERROR:   log.ErrorLevel,
	}
)
