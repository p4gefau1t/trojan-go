package log

import (
	"os"
)

//LogLevel how much log to dump
//0: ALL; 1: INFO; 2: WARN; 3: ERROR; 4: FATAL; 5: OFF
type LogLevel int

const (
	All        LogLevel = 0
	InfoLevel  LogLevel = 1
	WarnLevel  LogLevel = 2
	ErrorLevel LogLevel = 3
	FatalLevel LogLevel = 4
	Off        LogLevel = 5
)

type Logger interface {
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
	Warn(v ...interface{})
	Warnf(format string, v ...interface{})
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
	Trace(v ...interface{})
	Tracef(format string, v ...interface{})
	SetLogLevel(level LogLevel)
}

var logger Logger = &EmptyLogger{}

type EmptyLogger struct{}

func (l *EmptyLogger) SetLogLevel(LogLevel) {}

func (l *EmptyLogger) Fatal(v ...interface{}) { os.Exit(1) }

func (l *EmptyLogger) Fatalf(format string, v ...interface{}) { os.Exit(1) }

// Error print error message to output
func (l *EmptyLogger) Error(v ...interface{}) {}

func (l *EmptyLogger) Errorf(format string, v ...interface{}) {}

func (l *EmptyLogger) Warn(v ...interface{}) {}

func (l *EmptyLogger) Warnf(format string, v ...interface{}) {}

func (l *EmptyLogger) Info(v ...interface{}) {}

func (l *EmptyLogger) Infof(format string, v ...interface{}) {}

func (l *EmptyLogger) Debug(v ...interface{}) {}

func (l *EmptyLogger) Debugf(format string, v ...interface{}) {}

func (l *EmptyLogger) Trace(v ...interface{}) {}

func (l *EmptyLogger) Tracef(format string, v ...interface{}) {}

func Error(v ...interface{}) {
	logger.Error(v...)
}

func Errorf(format string, v ...interface{}) {
	logger.Errorf(format, v...)
}

func Warn(v ...interface{}) {
	logger.Warn(v...)
}

func Warnf(format string, v ...interface{}) {
	logger.Warnf(format, v...)
}

func Info(v ...interface{}) {
	logger.Info(v...)
}

func Infof(format string, v ...interface{}) {
	logger.Warnf(format, v...)
}

func Debug(v ...interface{}) {
	logger.Debug(v...)
}

func Debugf(format string, v ...interface{}) {
	logger.Warnf(format, v...)
}

func Trace(v ...interface{}) {
	logger.Trace(v...)
}

func Tracef(format string, v ...interface{}) {
	logger.Tracef(format, v...)
}

func Fatal(v ...interface{}) {
	logger.Fatal(v...)
}

func Fatalf(format string, v ...interface{}) {
	logger.Fatalf(format, v...)
}

func SetLogLevel(level LogLevel) {
	logger.SetLogLevel(level)
}

func RegisterLogger(l Logger) {
	logger = l
}
