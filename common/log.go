package log

import (
	"io"
	"os"
	"sync"

	goslog "golang.org/x/exp/slog"
)

var (
	// The globally available logger
	Logger *goslog.Logger
	// The child loggers for components
	loggers      = make(map[string]*goslog.Logger)
	loggerMutext sync.Mutex
	opts         goslog.HandlerOptions
)

// init initializes the default logger with given logging level
// from the configuration.
func init() {
	// TODO getting configuration parameters of the control,
	// then use these parameters to customize the logger.
	opts.Level = goslog.LevelInfo

	Logger = goslog.New(goslog.NewJSONHandler(os.Stdout, &opts))
	goslog.SetDefault(Logger)
}

func GetLogger(componentName string, args ...any) *goslog.Logger {
	if len(componentName) == 0 {
		return Logger
	}

	loggerMutext.Lock()
	defer loggerMutext.Unlock()

	theLogger := loggers[componentName]
	if theLogger != nil {
		return theLogger
	}
	// there is no such logger exists yet, create one logger based on
	// the given arguments and place it into the loggers

	theLogger = Logger.With(goslog.Group(componentName, args...))
	loggers[componentName] = theLogger
	return theLogger
}

// AddWriter adds a writer to the logger output in addition to the
// stdout.
func AddWriter(writer *io.Writer) {
	loggerMutext.Lock()
	defer loggerMutext.Unlock()

	multiWriters := io.MultiWriter(os.Stdout, *writer)
	Logger = goslog.New(goslog.NewJSONHandler(multiWriters, &opts))
	goslog.SetDefault(Logger)
}

// UseWriter change the writer of the logger output
func UseWriter(writer *io.Writer) {
	loggerMutext.Lock()
	defer loggerMutext.Unlock()

	Logger = goslog.New(goslog.NewJSONHandler(*writer, &opts))
	goslog.SetDefault(Logger)
}

// ResetWriter set the writer of the logger output to stdout
func ResetWriter() {
	loggerMutext.Lock()
	defer loggerMutext.Unlock()

	Logger = goslog.New(goslog.NewJSONHandler(os.Stdout, &opts))
	goslog.SetDefault(Logger)
}
