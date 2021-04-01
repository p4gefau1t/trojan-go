// The colorful and simple logging library
// Copyright (c) 2017 Fadhli Dzil Ikram

package golog

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/log/golog/colorful"
)

func init() {
	log.RegisterLogger(New(os.Stdout))
}

// FdWriter interface extends existing io.Writer with file descriptor function
// support
type FdWriter interface {
	io.Writer
	Fd() uintptr
}

// Logger struct define the underlying storage for single logger
type Logger struct {
	mu        sync.RWMutex
	color     bool
	out       io.Writer
	debug     bool
	timestamp bool
	quiet     bool
	buf       colorful.ColorBuffer
	logLevel  int
}

// Prefix struct define plain and color byte
type Prefix struct {
	Plain []byte
	Color []byte
	File  bool
}

var (
	// Plain prefix template
	plainFatal = []byte("[FATAL] ")
	plainError = []byte("[ERROR] ")
	plainWarn  = []byte("[WARN]  ")
	plainInfo  = []byte("[INFO]  ")
	plainDebug = []byte("[DEBUG] ")
	plainTrace = []byte("[TRACE] ")

	// FatalPrefix show fatal prefix
	FatalPrefix = Prefix{
		Plain: plainFatal,
		Color: colorful.Red(plainFatal),
		File:  true,
	}

	// ErrorPrefix show error prefix
	ErrorPrefix = Prefix{
		Plain: plainError,
		Color: colorful.Red(plainError),
		File:  true,
	}

	// WarnPrefix show warn prefix
	WarnPrefix = Prefix{
		Plain: plainWarn,
		Color: colorful.Orange(plainWarn),
	}

	// InfoPrefix show info prefix
	InfoPrefix = Prefix{
		Plain: plainInfo,
		Color: colorful.Green(plainInfo),
	}

	// DebugPrefix show info prefix
	DebugPrefix = Prefix{
		Plain: plainDebug,
		Color: colorful.Purple(plainDebug),
		File:  true,
	}

	// TracePrefix show info prefix
	TracePrefix = Prefix{
		Plain: plainTrace,
		Color: colorful.Cyan(plainTrace),
	}
)

// New returns new Logger instance with predefined writer output and
// automatically detect terminal coloring support
func New(out FdWriter) *Logger {
	return &Logger{
		color:     terminal.IsTerminal(int(out.Fd())),
		out:       out,
		timestamp: true,
	}
}

func (l *Logger) SetLogLevel(level log.LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logLevel = int(level)
}

func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.color = false
	if fdw, ok := w.(FdWriter); ok {
		l.color = terminal.IsTerminal(int(fdw.Fd()))
	}
	l.out = w
}

// WithColor explicitly turn on colorful features on the log
func (l *Logger) WithColor() *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.color = true
	return l
}

// WithoutColor explicitly turn off colorful features on the log
func (l *Logger) WithoutColor() *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.color = false
	return l
}

// WithDebug turn on debugging output on the log to reveal debug and trace level
func (l *Logger) WithDebug() *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.debug = true
	return l
}

// WithoutDebug turn off debugging output on the log
func (l *Logger) WithoutDebug() *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.debug = false
	return l
}

// IsDebug check the state of debugging output
func (l *Logger) IsDebug() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.debug
}

// WithTimestamp turn on timestamp output on the log
func (l *Logger) WithTimestamp() *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.timestamp = true
	return l
}

// WithoutTimestamp turn off timestamp output on the log
func (l *Logger) WithoutTimestamp() *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.timestamp = false
	return l
}

// Quiet turn off all log output
func (l *Logger) Quiet() *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.quiet = true
	return l
}

// NoQuiet turn on all log output
func (l *Logger) NoQuiet() *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.quiet = false
	return l
}

// IsQuiet check for quiet state
func (l *Logger) IsQuiet() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.quiet
}

// Output print the actual value
func (l *Logger) Output(depth int, prefix Prefix, data string) error {
	// Check if quiet is requested, and try to return no error and be quiet
	if l.IsQuiet() {
		return nil
	}
	// Get current time
	now := time.Now()
	// Temporary storage for file and line tracing
	var file string
	var line int
	var fn string
	// Check if the specified prefix needs to be included with file logging
	if prefix.File {
		var ok bool
		var pc uintptr

		// Get the caller filename and line
		if pc, file, line, ok = runtime.Caller(depth + 2); !ok {
			file = "<unknown file>"
			fn = "<unknown function>"
			line = 0
		} else {
			file = filepath.Base(file)
			fn = runtime.FuncForPC(pc).Name()
		}
	}
	// Acquire exclusive access to the shared buffer
	l.mu.Lock()
	defer l.mu.Unlock()
	// Reset buffer so it start from the begining
	l.buf.Reset()
	// Write prefix to the buffer
	if l.color {
		l.buf.Append(prefix.Color)
	} else {
		l.buf.Append(prefix.Plain)
	}
	// Check if the log require timestamping
	if l.timestamp {
		// Print timestamp color if color enabled
		if l.color {
			l.buf.Blue()
		}
		// Print date and time
		year, month, day := now.Date()
		l.buf.AppendInt(year, 4)
		l.buf.AppendByte('/')
		l.buf.AppendInt(int(month), 2)
		l.buf.AppendByte('/')
		l.buf.AppendInt(day, 2)
		l.buf.AppendByte(' ')
		hour, min, sec := now.Clock()
		l.buf.AppendInt(hour, 2)
		l.buf.AppendByte(':')
		l.buf.AppendInt(min, 2)
		l.buf.AppendByte(':')
		l.buf.AppendInt(sec, 2)
		l.buf.AppendByte(' ')
		// Print reset color if color enabled
		if l.color {
			l.buf.Off()
		}
	}
	// Add caller filename and line if enabled
	if prefix.File {
		// Print color start if enabled
		if l.color {
			l.buf.Orange()
		}
		// Print filename and line
		l.buf.Append([]byte(fn))
		l.buf.AppendByte(':')
		l.buf.Append([]byte(file))
		l.buf.AppendByte(':')
		l.buf.AppendInt(line, 0)
		l.buf.AppendByte(' ')
		// Print color stop
		if l.color {
			l.buf.Off()
		}
	}
	// Print the actual string data from caller
	l.buf.Append([]byte(data))
	if len(data) == 0 || data[len(data)-1] != '\n' {
		l.buf.AppendByte('\n')
	}
	// Flush buffer to output
	_, err := l.out.Write(l.buf.Buffer)
	return err
}

// Fatal print fatal message to output and quit the application with status 1
func (l *Logger) Fatal(v ...interface{}) {
	if l.logLevel <= 4 {
		l.Output(1, FatalPrefix, fmt.Sprintln(v...))
	}
	os.Exit(1)
}

// Fatalf print formatted fatal message to output and quit the application
// with status 1
func (l *Logger) Fatalf(format string, v ...interface{}) {
	if l.logLevel <= 4 {
		l.Output(1, FatalPrefix, fmt.Sprintf(format, v...))
	}
	os.Exit(1)
}

// Error print error message to output
func (l *Logger) Error(v ...interface{}) {
	if l.logLevel <= 3 {
		l.Output(1, ErrorPrefix, fmt.Sprintln(v...))
	}
}

// Errorf print formatted error message to output
func (l *Logger) Errorf(format string, v ...interface{}) {
	if l.logLevel <= 3 {
		l.Output(1, ErrorPrefix, fmt.Sprintf(format, v...))
	}
}

// Warn print warning message to output
func (l *Logger) Warn(v ...interface{}) {
	if l.logLevel <= 2 {
		l.Output(1, WarnPrefix, fmt.Sprintln(v...))
	}
}

// Warnf print formatted warning message to output
func (l *Logger) Warnf(format string, v ...interface{}) {
	if l.logLevel <= 2 {
		l.Output(1, WarnPrefix, fmt.Sprintf(format, v...))
	}
}

// Info print informational message to output
func (l *Logger) Info(v ...interface{}) {
	if l.logLevel <= 1 {
		l.Output(1, InfoPrefix, fmt.Sprintln(v...))
	}
}

// Infof print formatted informational message to output
func (l *Logger) Infof(format string, v ...interface{}) {
	if l.logLevel <= 1 {
		l.Output(1, InfoPrefix, fmt.Sprintf(format, v...))
	}
}

// Debug print debug message to output if debug output enabled
func (l *Logger) Debug(v ...interface{}) {
	if l.logLevel == 0 {
		l.Output(1, DebugPrefix, fmt.Sprintln(v...))
	}
}

// Debugf print formatted debug message to output if debug output enabled
func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.logLevel == 0 {
		l.Output(1, DebugPrefix, fmt.Sprintf(format, v...))
	}
}

// Trace print trace message to output if debug output enabled
func (l *Logger) Trace(v ...interface{}) {
	if l.logLevel == 0 {
		l.Output(1, TracePrefix, fmt.Sprintln(v...))
	}
}

// Tracef print formatted trace message to output if debug output enabled
func (l *Logger) Tracef(format string, v ...interface{}) {
	if l.logLevel == 0 {
		l.Output(1, TracePrefix, fmt.Sprintf(format, v...))
	}
}
