package store

import "log"

var (
	logEnabled bool = false
)

type Level string

var CRIT Level = "CRIT"
var ERRO Level = "ERRO"
var INFO Level = "INFO"
var DEBU Level = "DEBU"

func SetLogEnable(enable bool) {
	logEnabled = enable
}

func logf(level string, fmt string, args ...interface{}) {
	if !logEnabled {
		return
	}
	log.Printf(level+" "+fmt, args...)
}

func infof(fmt string, args ...interface{}) {
	logf(string(INFO), fmt, args...)
}

func errf(fmt string, args ...interface{}) {
	logf(string(ERRO), fmt, args...)
}
