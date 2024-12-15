package wlog

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"path"
	"runtime"
)

type LoggerEntry interface {
	Field(key string, value interface{}) LoggerEntry
	Err(err error) LoggerEntry
	Log()
}

type loggerEntry struct {
	level   zapcore.Level
	logger  *zap.Logger
	message string
}

func NewLoggerEntry(level zapcore.Level, message string) LoggerEntry {
	return &loggerEntry{
		level:   level,
		logger:  logger,
		message: message,
	}
}

// 获取调用日志的函数或方法全名的最后一部分，一般来说是最后一个斜杠后的部分
// 对于函数，其全名为 module/paths/pkg.FuncName(其中 paths 是从模块根目录到包的相对路径)，处理后返回的是 pkg.FuncName
// 对于方法，其全名同理于函数，处理后返回的是 pkg.(*Type).MethodName 或者 pkg.Type.MethodName
func CallerName() string {
	pc, _, _, ok := runtime.Caller(2)
	if ok {
		fullName := runtime.FuncForPC(pc).Name()
		baseName := path.Base(fullName)
		return baseName
	} else {
		return "unknown"
	}
}

func (l *loggerEntry) Field(key string, value interface{}) LoggerEntry {
	l.logger = l.logger.With(zap.Any(key, value))
	return l
}

func (l *loggerEntry) Err(err error) LoggerEntry {
	l.logger = l.logger.With(zap.Error(err))
	return l
}

func (l *loggerEntry) Log() {
	l.logger.With(zap.String("caller", CallerName())).
		WithOptions(zap.AddCallerSkip(1)).Check(l.level, l.message).Write()
}
