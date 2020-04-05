package simplelog

import (
	golog "log"

	"github.com/p4gefau1t/trojan-go/log"
)

func init() {
	log.RegisterLogger(&SimpleLogger{})
}

type SimpleLogger struct {
	logLevel log.LogLevel
}

func (l *SimpleLogger) SetLogLevel(level log.LogLevel) {
	l.logLevel = level
}

func (l *SimpleLogger) Fatal(v ...interface{}) {
	golog.Fatal(v...)
}

func (l *SimpleLogger) Fatalf(format string, v ...interface{}) {
	golog.Fatalf(format, v...)
}

func (l *SimpleLogger) Error(v ...interface{}) {
	golog.Println(v...)
}

func (l *SimpleLogger) Errorf(format string, v ...interface{}) {
	golog.Printf(format, v...)
}

func (l *SimpleLogger) Warn(v ...interface{}) {
	golog.Println(v...)
}

func (l *SimpleLogger) Warnf(format string, v ...interface{}) {
	golog.Printf(format, v...)
}

func (l *SimpleLogger) Info(v ...interface{}) {
	golog.Println(v...)
}

func (l *SimpleLogger) Infof(format string, v ...interface{}) {
	golog.Printf(format, v...)
}

func (l *SimpleLogger) Debug(v ...interface{}) {
	golog.Println(v...)
}

func (l *SimpleLogger) Debugf(format string, v ...interface{}) {
	golog.Printf(format, v...)
}

func (l *SimpleLogger) Trace(v ...interface{}) {
	golog.Println(v...)
}

func (l *SimpleLogger) Tracef(format string, v ...interface{}) {
	golog.Printf(format, v...)
}
