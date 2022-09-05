package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var (
	debugEnabled bool
	debugFile    *os.File
)

func EnableDebugging() {
	debugEnabled = true
	logFile := filepath.Join(os.TempDir(), "lazytrivy-logger.log")
	debugFile, _ = os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
}

func Debug(format string, args ...interface{}) {
	log("DEBUG", format, args...)
}

func Error(format string, args ...interface{}) {
	log("ERROR", format, args...)
}

func WithError(err error, format string, args ...interface{}) {
	Error(format, args...)
	log("\t", "Error: %s", err)
}

func log(level string, format string, args ...interface{}) {
	if debugEnabled {
		_, _ = fmt.Fprintf(debugFile, fmt.Sprintf("%s [%s] ", time.RFC3339, level)+fmt.Sprintf(format, args...))
		_, _ = fmt.Fprintln(debugFile)
	}
}
