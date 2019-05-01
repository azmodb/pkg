// Package log implements a simple logging package. It provides functions
// Debug, Info, Error, Fatal, Panic plus formatting variants such as
// Infof.
package log

import (
	"log"
	"os"
	"sync"
)

const defLogFlags = log.Ldate | log.Ltime | log.LUTC | log.Lmicroseconds

// Logger is the interface for logging messages to standard error.
type Logger interface {
	// Printf writes a formated message to the log.
	Printf(format string, args ...interface{})

	// Print writes a message to the log.
	Print(args ...interface{})

	// Fatal writes a message to the log and aborts.
	Fatal(args ...interface{})

	// Fatalf writes a formated message to the log and aborts.
	Fatalf(format string, args ...interface{})

	// Panic is equivalent to Print() followed by a call to panic().
	Panic(args ...interface{})

	// Panicf is equivalent to Printf() followed by a call to panic().
	Panicf(format string, args ...interface{})
}

// Level represents the level of logging.
type Level int

// Different levels of logging.
const (
	DebugLevel Level = iota
	InfoLevel
	ErrorLevel
	DisabledLevel
)

type state struct {
	sync.RWMutex
	level Level
}

var global *state

func init() { global = &state{} }

// SetLevel sets the current level of logging.
func SetLevel(level Level) {
	global.Lock()
	global.level = level
	global.Unlock()
}

// getLevel returns the current logging level.
func getLevel() Level {
	global.RLock()
	level := global.level
	global.RUnlock()
	return level
}

// Debugf log to the debug logs. Arguments are handled in the manner
// of fmt.Printf; a newline is appended if missing.
func Debugf(format string, args ...interface{}) {
	debugLog.Printf(format, args...)
}

// Debug log to the debug logs. Arguments are handled in the manner
// of fmt.Print; a newline is appended if missing.
func Debug(args ...interface{}) {
	debugLog.Print(args...)
}

// Infof log to the info logs. Arguments are handled in the manner
// of fmt.Printf; a newline is appended if missing.
func Infof(format string, args ...interface{}) {
	infoLog.Printf(format, args...)
}

// Info log to the info logs. Arguments are handled in the manner
// of fmt.Print; a newline is appended if missing.
func Info(args ...interface{}) {
	infoLog.Print(args...)
}

// Errorf log to the error logs. Arguments are handled in the manner
// of fmt.Printf; a newline is appended if missing.
func Errorf(format string, args ...interface{}) {
	errorLog.Printf(format, args...)
}

// Error log to the error logs. Arguments are handled in the manner
// of fmt.Print; a newline is appended if missing.
func Error(args ...interface{}) {
	errorLog.Print(args...)
}

// Fatalf log to the fatal logs, regardless of the current log level.
// Arguments are handled in the manner of fmt.Printf; a newline is
// appended if missing.
func Fatalf(format string, args ...interface{}) {
	fatalLog.Fatalf(format, args...)
}

// Fatal log to the fatal logs, regardless of the current log level.
// Arguments are handled in the manner of fmt.Print; a newline is
// appended if missing.
func Fatal(args ...interface{}) {
	fatalLog.Fatal(args...)
}

// Panicf log to the panic logs, regardless of the current log level.
// Arguments are handled in the manner of fmt.Printf; a newline is
// appended if missing.
func Panicf(format string, args ...interface{}) {
	fatalLog.Panicf(format, args...)
}

// Panic log to the panic logs, regardless of the current log level.
// Arguments are handled in the manner of fmt.Print; a newline is
// appended if missing.
func Panic(args ...interface{}) {
	fatalLog.Panic(args...)
}

var _ Logger = (*logger)(nil)

func newStdLogger(prefix string) *log.Logger {
	return log.New(os.Stderr, prefix, defLogFlags)
}

// Default loggers for each log level.
var (
	debugLog = &logger{newStdLogger("DEBUG "), DebugLevel}
	infoLog  = &logger{newStdLogger("INFO  "), InfoLevel}
	errorLog = &logger{newStdLogger("ERROR "), ErrorLevel}
	fatalLog = &logger{newStdLogger("FATAL "), DisabledLevel}
)

type logger struct {
	log   Logger
	level Level
}

// New creates a new level logger.
func New(log Logger, level Level) Logger {
	return &logger{
		level: level,
		log:   log,
	}
}

func (l *logger) Printf(format string, args ...interface{}) {
	level := getLevel()
	if l.level >= level {
		l.log.Printf(format, args...)
	}
}

func (l *logger) Print(args ...interface{}) {
	level := getLevel()
	if l.level >= level {
		l.log.Print(args...)
	}
}

func (l *logger) Fatal(args ...interface{}) {
	l.log.Fatal(args...)
}

func (l *logger) Fatalf(format string, args ...interface{}) {
	l.log.Fatalf(format, args...)
}

func (l *logger) Panic(args ...interface{}) {
	l.log.Panic(args...)
}

func (l *logger) Panicf(format string, args ...interface{}) {
	l.log.Panicf(format, args...)
}
