package logger

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

func Logf(level string, fmt string, args ...interface{}) {
	if !logEnabled {
		return
	}
	log.Printf(level+" "+fmt, args...)
}

func Infof(fmt string, args ...interface{}) {
	Logf(string(INFO), fmt, args...)
}

func Errf(fmt string, args ...interface{}) {
	Logf(string(ERRO), fmt, args...)
}
