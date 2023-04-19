package log

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"sql_zh_exporter/config"
	"strings"

	"os"
	"runtime"
	"strconv"
	"time"
)

const (
	traceId = "traceid"

	logLevel = "log.level"
	logWrite = "log.writer"

	levelInfo  = "INFO"
	levelDebug = "DEBUG"
	levelWarn  = "WARN"
	levelError = "ERROR"
)

var console string

func Init(confileFile string) {

	console = config.GetConfig(logWrite, confileFile)

	zerolog.TimeFieldFormat = time.RFC3339Nano
	var level zerolog.Level
	switch config.GetConfig(logLevel, confileFile) {
	case levelDebug:
		level = zerolog.DebugLevel
	case levelWarn:
		level = zerolog.WarnLevel
	case levelError:
		level = zerolog.ErrorLevel
	default:
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	switch config.GetConfig(logWrite, confileFile) {
	default:
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	}
}

func useConsoleWrite() bool {
	return strings.EqualFold("CONSOLE", console)
}

func Panic() *zerolog.Event {

	_, file, line, ok := runtime.Caller(1)
	e := log.Panic()
	if ok {
		if useConsoleWrite() {
			e = e.Str(zerolog.CallerFieldName, file+":"+strconv.Itoa(line))
		} else {
			e = e.Str("line", file+":"+strconv.Itoa(line))
		}
	}
	return e
}
func Error() *zerolog.Event {

	_, file, line, ok := runtime.Caller(1)
	e := log.Error()
	if ok {
		if useConsoleWrite() {
			e = e.Str(zerolog.CallerFieldName, file+":"+strconv.Itoa(line))
		} else {
			e = e.Str("line", file+":"+strconv.Itoa(line))
		}
	}
	return e
}

func Debug() *zerolog.Event {
	_, file, line, ok := runtime.Caller(1)
	e := log.Debug()
	if ok {
		if useConsoleWrite() {
			e = e.Str(zerolog.CallerFieldName, file+":"+strconv.Itoa(line))
		} else {
			e = e.Str("line", file+":"+strconv.Itoa(line))
		}
	}
	return e
}

func Warn() *zerolog.Event {
	_, file, line, ok := runtime.Caller(1)
	e := log.Warn()
	if ok {
		if useConsoleWrite() {
			e = e.Str(zerolog.CallerFieldName, file+":"+strconv.Itoa(line))
		} else {
			e = e.Str("line", file+":"+strconv.Itoa(line))
		}
	}
	return e
}

func Info() *zerolog.Event {
	_, file, line, ok := runtime.Caller(1)
	e := log.Info()
	if ok {
		if useConsoleWrite() {
			e = e.Str(zerolog.CallerFieldName, file+":"+strconv.Itoa(line))
		} else {
			e = e.Str("line", file+":"+strconv.Itoa(line))
		}

	}
	return e
}
