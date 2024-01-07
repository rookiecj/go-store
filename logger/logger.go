package logger

import "log"

var (
	logEnabled bool = false
)

type Level string

var CRIT Level = "CRIT"
var ERROR Level = "ERROR"
var INFO Level = "INFO"
var DEBUG Level = "DEBU"

func SetLogEnable(enable bool) {
	logEnabled = enable
}

func LogForcedf(fmt string, args ...interface{}) {
	log.Printf(fmt, args...)
}

func Logf(level string, fmt string, args ...interface{}) {
	if !logEnabled {
		return
	}
	log.Printf(level+" "+fmt, args...)
}

func Debugf(fmt string, args ...interface{}) {
	Logf(string(DEBUG), fmt, args...)
}

func Infof(fmt string, args ...interface{}) {
	Logf(string(INFO), fmt, args...)
}

func Errf(fmt string, args ...interface{}) {
	Logf(string(ERROR), fmt, args...)
}

func Critf(fmt string, args ...interface{}) {
	Logf(string(CRIT), fmt, args...)
}
