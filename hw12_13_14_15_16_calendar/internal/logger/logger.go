package logger

import (
	"io"
	"log"
	"os"
	"strings"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Logger struct {
	level    Level
	debugLog *log.Logger
	infoLog  *log.Logger
	warnLog  *log.Logger
	errorLog *log.Logger
}

func parseLevel(level string) Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARN", "WARNING":
		return LevelWarn
	case "ERROR":
		return LevelError
	default:
		return LevelInfo
	}
}

func New(level string) *Logger {
	parsedLevel := parseLevel(level)

	//  побитовое ИЛИ: дата + время + микросекунды.
	flags := log.Ldate | log.Ltime | log.Lmicroseconds

	return &Logger{
		level:    parsedLevel,
		debugLog: log.New(os.Stdout, "DEBUG: ", flags),
		infoLog:  log.New(os.Stdout, "INFO: ", flags),
		warnLog:  log.New(os.Stdout, "WARN: ", flags),
		errorLog: log.New(os.Stderr, "ERROR: ", flags),
	}
}

func NewWithWriter(level string, out io.Writer, errOut io.Writer) *Logger {
	parsedLevel := parseLevel(level)

	flags := log.Ldate | log.Ltime | log.Lmicroseconds

	return &Logger{
		level:    parsedLevel,
		debugLog: log.New(out, "DEBUG: ", flags),
		infoLog:  log.New(out, "INFO: ", flags),
		warnLog:  log.New(out, "WARN: ", flags),
		errorLog: log.New(errOut, "ERROR: ", flags),
	}
}

func (l *Logger) Debug(msg string) {
	if l.level <= LevelDebug {
		l.debugLog.Println(msg)
	}
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	if l.level <= LevelDebug {
		l.debugLog.Printf(format, args...)
	}
}

func (l *Logger) Info(msg string) {
	if l.level <= LevelInfo {
		l.infoLog.Println(msg)
	}
}

func (l *Logger) Infof(format string, args ...interface{}) {
	if l.level <= LevelInfo {
		l.infoLog.Printf(format, args...)
	}
}

func (l *Logger) Warn(msg string) {
	if l.level <= LevelWarn {
		l.warnLog.Println(msg)
	}
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	if l.level <= LevelWarn {
		l.warnLog.Printf(format, args...)
	}
}

func (l *Logger) Error(msg string) {
	if l.level <= LevelError {
		l.errorLog.Println(msg)
	}
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	if l.level <= LevelError {
		l.errorLog.Printf(format, args...)
	}
}

func NewFileLogger(level string, filepath string) (*Logger, error) {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		return nil, err
	}

	parsedLevel := parseLevel(level)
	flags := log.Ldate | log.Ltime | log.Lmicroseconds

	return &Logger{
		level:    parsedLevel,
		debugLog: log.New(file, "DEBUG: ", flags),
		infoLog:  log.New(file, "INFO: ", flags),
		warnLog:  log.New(file, "WARN: ", flags),
		errorLog: log.New(file, "ERROR: ", flags),
	}, nil
}
