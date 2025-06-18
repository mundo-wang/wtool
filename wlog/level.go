package wlog

import (
	"fmt"
	"go.uber.org/zap/zapcore"
)

func Debug(message string) LoggerEntry {
	return newLoggerEntry(zapcore.DebugLevel, message)
}

func Info(message string) LoggerEntry {
	return newLoggerEntry(zapcore.InfoLevel, message)
}

func Warn(message string) LoggerEntry {
	return newLoggerEntry(zapcore.WarnLevel, message)
}

func Error(message string) LoggerEntry {
	return newLoggerEntry(zapcore.ErrorLevel, message)
}

func Fatal(message string) LoggerEntry {
	return newLoggerEntry(zapcore.FatalLevel, message)
}

func Panic(message string) LoggerEntry {
	return newLoggerEntry(zapcore.PanicLevel, message)
}

func Debugf(format string, args ...interface{}) LoggerEntry {
	message := fmt.Sprintf(format, args...)
	return newLoggerEntry(zapcore.DebugLevel, message)
}

func Infof(format string, args ...interface{}) LoggerEntry {
	message := fmt.Sprintf(format, args...)
	return newLoggerEntry(zapcore.InfoLevel, message)
}

func Warnf(format string, args ...interface{}) LoggerEntry {
	message := fmt.Sprintf(format, args...)
	return newLoggerEntry(zapcore.WarnLevel, message)
}

func Errorf(format string, args ...interface{}) LoggerEntry {
	message := fmt.Sprintf(format, args...)
	return newLoggerEntry(zapcore.ErrorLevel, message)
}

func Fatalf(format string, args ...interface{}) LoggerEntry {
	message := fmt.Sprintf(format, args...)
	return newLoggerEntry(zapcore.FatalLevel, message)
}

func Panicf(format string, args ...interface{}) LoggerEntry {
	message := fmt.Sprintf(format, args...)
	return newLoggerEntry(zapcore.PanicLevel, message)
}
