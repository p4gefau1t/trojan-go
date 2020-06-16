package simplelog

import (
	"io"
	golog "log"
	"os"

	"github.com/p4gefau1t/trojan-go/log"
	_ "github.com/p4gefau1t/trojan-go/log/tz"
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
	if l.logLevel <= log.FatalLevel {
		golog.Fatal(v...)
	}
	os.Exit(1)
}

func (l *SimpleLogger) Fatalf(format string, v ...interface{}) {
	if l.logLevel <= log.FatalLevel {
		golog.Fatalf(format, v...)
	}
	os.Exit(1)
}

func (l *SimpleLogger) Error(v ...interface{}) {
	if l.logLevel <= log.ErrorLevel {
		golog.Println(v...)
	}
}

func (l *SimpleLogger) Errorf(format string, v ...interface{}) {
	if l.logLevel <= log.ErrorLevel {
		golog.Printf(format, v...)
	}
}

func (l *SimpleLogger) Warn(v ...interface{}) {
	if l.logLevel <= log.WarnLevel {
		golog.Println(v...)
	}
}

func (l *SimpleLogger) Warnf(format string, v ...interface{}) {
	if l.logLevel <= log.WarnLevel {
		golog.Printf(format, v...)
	}
}

func (l *SimpleLogger) Info(v ...interface{}) {
	if l.logLevel <= log.InfoLevel {
		golog.Println(v...)
	}
}

func (l *SimpleLogger) Infof(format string, v ...interface{}) {
	if l.logLevel <= log.InfoLevel {
		golog.Printf(format, v...)
	}
}

func (l *SimpleLogger) Debug(v ...interface{}) {
	if l.logLevel <= log.AllLevel {
		golog.Println(v...)
	}
}

func (l *SimpleLogger) Debugf(format string, v ...interface{}) {
	if l.logLevel <= log.AllLevel {
		golog.Printf(format, v...)
	}
}

func (l *SimpleLogger) Trace(v ...interface{}) {
	if l.logLevel <= log.AllLevel {
		golog.Println(v...)
	}
}

func (l *SimpleLogger) Tracef(format string, v ...interface{}) {
	if l.logLevel <= log.AllLevel {
		golog.Printf(format, v...)
	}
}

func (l *SimpleLogger) SetOutput(io.Writer) {
	//do nothing
}
