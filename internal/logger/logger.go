package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelError
)

var (
	currentLevel Level = LevelInfo
	logFile      *os.File
	mu           sync.Mutex
	sysLogger    *log.Logger
)

// Init initializes the application logger with a level and output file.
// Valid levels: "debug", "info", "error"
func Init(levelStr, filename string) error {
	mu.Lock()
	defer mu.Unlock()

	// Parse level
	switch strings.ToLower(levelStr) {
	case "debug":
		currentLevel = LevelDebug
	case "error":
		currentLevel = LevelError
	default:
		currentLevel = LevelInfo // default
	}

	// Open file
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %v", filename, err)
	}
	logFile = f

	// MultiWriter to both stdout and file
	mw := io.MultiWriter(os.Stdout, logFile)
	sysLogger = log.New(mw, "", log.LstdFlags)

	return nil
}

// Close closes the underlying log file
func Close() {
	mu.Lock()
	defer mu.Unlock()
	if logFile != nil {
		logFile.Close()
	}
}

// Debug logs debug messages
func Debug(format string, v ...interface{}) {
	if currentLevel <= LevelDebug {
		sysLogger.Output(2, "[DEBUG] "+fmt.Sprintf(format, v...))
	}
}

// Info logs standard information
func Info(format string, v ...interface{}) {
	if currentLevel <= LevelInfo {
		sysLogger.Output(2, "[INFO]  "+fmt.Sprintf(format, v...))
	}
}

// Error logs error messages
func Error(format string, v ...interface{}) {
	if currentLevel <= LevelError {
		sysLogger.Output(2, "[ERROR] "+fmt.Sprintf(format, v...))
	}
}

// Fatal logs an error and exits
func Fatal(format string, v ...interface{}) {
	sysLogger.Output(2, "[FATAL] "+fmt.Sprintf(format, v...))
	Close()
	os.Exit(1)
}
