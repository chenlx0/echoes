package lg

import (
	"fmt"
	"log"
	"os"
)

// Custom log constants
const (
	INFO            = LogLevel(1)
	DEBUG           = LogLevel(2)
	WARN            = LogLevel(3)
	ERROR           = LogLevel(4)
	FATAL           = LogLevel(5)
	DefaultLogDepth = 3
)

// AppLogFunc log function for http handler
type AppLogFunc func(l LogLevel, f string, args ...interface{})

// Logger an instance for log output
type Logger interface {
	Output(depth int, s string) error
}

// LogLevel custom log level type
type LogLevel uint8

func (l *LogLevel) String() string {
	switch *l {
	case INFO:
		return "INFO"
	case DEBUG:
		return "DEBUG"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	}
	return "INVALID_LOG_TYPE"
}

// Logf format and print log via Logger.Output method
func Logf(logger Logger, cfgLevel LogLevel, level LogLevel, f string, args ...interface{}) {
	if cfgLevel > level {
		return
	}
	msg := fmt.Sprintf("[%s] %s %v", level.String(), f, args)
	logger.Output(DefaultLogDepth, msg)
}

// LogFatal call for crash event before server established
func LogFatal(f string, args ...interface{}) {
	logger := log.New(os.Stderr, "SERVER-FATAL:", log.Ldate|log.Ltime|log.Lmicroseconds)
	Logf(logger, FATAL, FATAL, f, args)
	os.Exit(1)
}

// GetAppLogFunc decorate and return a AppLogFunc
func GetAppLogFunc(file *os.File, cfgLevel LogLevel) AppLogFunc {
	logger := log.New(file, "", log.Ldate|log.Ltime|log.Lmicroseconds)
	appLogFunc := func(l LogLevel, f string, args ...interface{}) {
		Logf(logger, cfgLevel, l, f, args)
	}
	return appLogFunc
}
