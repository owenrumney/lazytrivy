package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var (
	debugEnabled bool
	traceEnabled bool
	debugFile    *os.File
)

func Configure() {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "/"
	}
	logDir := filepath.Join(home, ".lazytrivy", "logs")
	_ = os.MkdirAll(logDir, os.ModePerm)

	logFile := filepath.Join(logDir, "lazytrivy.log")
	debugFile, _ = os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600) // nolint: nosnakecase
}

func EnableDebugging() {
	debugEnabled = true
}

func EnableTracing() {
	traceEnabled = true
}

func Tracef(format string, args ...interface{}) {
	if traceEnabled {
		logf("TRACE", format, args...)
	}

}

func Debugf(format string, args ...interface{}) {
	if debugEnabled {
		logf("DEBUG", format, args...)
	}
}

func Infof(format string, args ...interface{}) {
	logf("INFO", format, args...)
}

func Errorf(format string, args ...interface{}) {
	logf("ERROR", format, args...)
}

func logf(level string, format string, args ...interface{}) {
	_, _ = fmt.Fprintf(debugFile, fmt.Sprintf("%s\t[%s]\t", time.Now().Format(time.RFC3339), level)+fmt.Sprintf(format, args...))
	_, _ = fmt.Fprintln(debugFile)
}
