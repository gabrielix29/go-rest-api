package log

import (
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

// Level type
type Level string

const (
	// LevelDebug usually only enabled when debugging
	LevelDebug Level = "debug"
	// LevelInfo general operational entries about what's going on inside the application
	LevelInfo Level = "info"
	// LevelWarn non-critical entries that deserve eyes
	LevelWarn Level = "warn"
	// LevelError used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	LevelError Level = "error"
	// LevelFatal Logs and then calls `logger.Exit(1)`
	LevelFatal Level = "fatal"
	// LevelPanic logs and then calls panic
	LevelPanic Level = "panic"
)

// String return string value of a Level constant
func (l *Level) String() string {
	return string(*l)
}

// ParseLevel takes a string level and returns the Level constant
func ParseLevel(level string) (Level, error) {
	switch level {
	case "debug":
		return LevelDebug, nil
	case "info":
		return LevelInfo, nil
	case "warn":
		return LevelWarn, nil
	case "error":
		return LevelError, nil
	case "fatal":
		return LevelFatal, nil
	case "panic":
		return LevelPanic, nil
	default:
		return "", fmt.Errorf("%v: %s", ErrUnknownLevel, level)
	}
}

// Logger implementation is responsible for providing structured and leveled
// logging functions.
// Logger implementation is responsible for providing structured and levled
// logging functions.
type Logger interface {
	Debug(args ...interface{})
	Debugln(args ...interface{})
	Debugf(msg string, args ...interface{})
	Infof(msg string, args ...interface{})
	Info(msg string)
	Infoln(...interface{})
	Warn(msg string)
	Warnln(...interface{})
	Warnf(msg string, args ...interface{})
	Error(msg string)
	Errorf(msg string, args ...interface{})
	Fatalf(msg string, args ...interface{})
	Print(args ...interface{})
	Printf(msg string, args ...interface{})
	Println(...interface{})
	Trace(args ...interface{})
	Tracef(msg string, args ...interface{})
	Traceln(...interface{})
	Verbose() bool

	// WithFields should return a logger which is annotated with the given
	// fields. These fields should be added to every logging call on the
	// returned logger.
	WithFields(m map[string]interface{}) Logger
	WithPrefix(prefix string) Logger

	Level() Level
}

// Fields own declaration of logrus Fields
type Fields logrus.Fields

// New returns a logger implemented using the logrus package.
func New(wr io.Writer, level Level, file string) Logger {
	if wr == nil {
		wr = os.Stderr
	}

	lg := logrus.New()
	lg.Out = wr

	lvl, err := logrus.ParseLevel(level.String())
	if err != nil {
		lvl = logrus.WarnLevel
		lg.Warnf("failed to parse log-level '%s', defaulting to 'warning'", level)
	}
	lg.SetLevel(lvl)
	lg.SetFormatter(getFormatter(false))

	if file != "" {
		fileHook, err := NewLogrusFileHook(file, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
		if err == nil {
			lg.Hooks.Add(fileHook)
		} else {
			lg.Warnf("Failed to open logfile, using standard out: %v", err)
		}
	}

	return &logrusLogger{
		Entry: logrus.NewEntry(lg),
	}
}

// logrusLogger provides functions for structured logging.
type logrusLogger struct {
	*logrus.Entry
}

// Level returns the Level that set on the Logger
func (l *logrusLogger) Level() Level {
	level, _ := ParseLevel(l.Entry.Logger.Level.String())
	return level
}

// WithFields should return a logger which is annotated with the given fields
func (l *logrusLogger) WithFields(fields map[string]interface{}) Logger {
	annotatedEntry := l.Entry.WithFields(fields)
	return &logrusLogger{
		Entry: annotatedEntry,
	}
}

// WithPrefix should return a logger which is annotated with the given prefix
func (l *logrusLogger) WithPrefix(prefix string) Logger {
	return l.WithFields(Fields{"prefix": prefix})
}

func (ll *logrusLogger) Error(msg string) {
	ll.Errorf(msg)
}

func (ll *logrusLogger) Info(msg string) {
	ll.Infof(msg)
}

func (ll *logrusLogger) Print(args ...interface{}) {
	ll.Debug(args...)
}

func (ll *logrusLogger) Warn(msg string) {
	ll.Warnf(msg)
}

func (ll *logrusLogger) Verbose() bool {
	return ll.Entry.Logger.GetLevel().String() == "debug"
}

// getFormatter returns the default log formatter.
func getFormatter(disableColors bool) *textFormatter {
	return &textFormatter{
		DisableColors:    disableColors,
		ForceFormatting:  true,
		ForceColors:      true,
		DisableTimestamp: false,
		FullTimestamp:    true,
		DisableSorting:   true,
		TimestampFormat:  "2006-01-02 15:04:05.000000",
		SpacePadding:     45,
	}
}
