package simplelog

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/p4gefau1t/trojan-go/log"
)

func init() {
	log.RegisterLogger(&SimpleLogger{
		out:    os.Stderr,
	})
}

type SimpleLogger struct {
	mu       sync.Mutex
	logLevel log.LogLevel
	buf      []byte
	out      io.Writer
}

func (l *SimpleLogger) SetLogLevel(level log.LogLevel) {
	l.logLevel = level
}

func (l *SimpleLogger) Fatal(v ...interface{}) {
	if l.logLevel <= log.FatalLevel {
		l.Output(2, fmt.Sprint(v...))
	}
	os.Exit(1)
}

func (l *SimpleLogger) Fatalf(format string, v ...interface{}) {
	if l.logLevel <= log.FatalLevel {
		l.Output(2, fmt.Sprintf(format, v...))
	}
	os.Exit(1)
}

func (l *SimpleLogger) Error(v ...interface{}) {
	if l.logLevel <= log.ErrorLevel {
		l.Output(2, fmt.Sprintln(v...))
	}
}

func (l *SimpleLogger) Errorf(format string, v ...interface{}) {
	if l.logLevel <= log.ErrorLevel {
		l.Output(2, fmt.Sprintf(format, v...))
	}
}

func (l *SimpleLogger) Warn(v ...interface{}) {
	if l.logLevel <= log.WarnLevel {
		l.Output(2, fmt.Sprintln(v...))
	}
}

func (l *SimpleLogger) Warnf(format string, v ...interface{}) {
	if l.logLevel <= log.WarnLevel {
		l.Output(2, fmt.Sprintf(format, v...))
	}
}

func (l *SimpleLogger) Info(v ...interface{}) {
	if l.logLevel <= log.InfoLevel {
		l.Output(2, fmt.Sprintln(v...))
	}
}

func (l *SimpleLogger) Infof(format string, v ...interface{}) {
	if l.logLevel <= log.InfoLevel {
		l.Output(2, fmt.Sprintf(format, v...))
	}
}

func (l *SimpleLogger) Debug(v ...interface{}) {
	if l.logLevel <= log.AllLevel {
		l.Output(2, fmt.Sprintln(v...))
	}
}

func (l *SimpleLogger) Debugf(format string, v ...interface{}) {
	if l.logLevel <= log.AllLevel {
		l.Output(2, fmt.Sprintf(format, v...))
	}
}

func (l *SimpleLogger) Trace(v ...interface{}) {
	if l.logLevel <= log.AllLevel {
		l.Output(2, fmt.Sprintln(v...))
	}
}

func (l *SimpleLogger) Tracef(format string, v ...interface{}) {
	if l.logLevel <= log.AllLevel {
		l.Output(2, fmt.Sprintf(format, v...))
	}
}

func (l *SimpleLogger) SetOutput(io.Writer) {
	//do nothing
}

func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

func (l *SimpleLogger) formatHeader(buf *[]byte, t time.Time, file string, fn string, line int) {
	year, month, day := t.Date()
	itoa(buf, year, 4)
	*buf = append(*buf, '/')
	itoa(buf, int(month), 2)
	*buf = append(*buf, '/')
	itoa(buf, day, 2)
	*buf = append(*buf, ' ')

	hour, min, sec := t.Clock()
	itoa(buf, hour, 2)
	*buf = append(*buf, ':')
	itoa(buf, min, 2)
	*buf = append(*buf, ':')
	itoa(buf, sec, 2)
	*buf = append(*buf, ' ')

	// Print filename and line
	*buf = append(*buf, fn...)
	*buf = append(*buf, ':')
	*buf = append(*buf, file...)
	*buf = append(*buf, ':')
	itoa(buf, line, -1)
	*buf = append(*buf, ": "...)

}

func (l *SimpleLogger) Output(calldepth int, s string) error {
	now := time.Now().Local() // force local TZ.
	var file string
	var fn string
	var line int
	l.mu.Lock()
	defer l.mu.Unlock()
	l.mu.Unlock()
	var ok bool
	var pc uintptr
	// Get the caller filename and line
	if pc, file, line, ok = runtime.Caller(calldepth + 2); !ok {
		file = "/<?>"
		fn = "<?>"
		line = 0
	} else {
		file = filepath.Base(file)
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		fn = runtime.FuncForPC(pc).Name()
		short = fn
		for i := len(fn) - 1; i > 0; i-- {
			if fn[i] == '/' {
				short = fn[i+1:]
				calldepth--
				if calldepth < 1 {
					break
				}
			}
		}
		fn = short
	}
	l.mu.Lock()
	l.buf = l.buf[:0]
	l.formatHeader(&l.buf, now, file, fn, line)
	l.buf = append(l.buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}
	_, err := l.out.Write(l.buf)
	return err
}

