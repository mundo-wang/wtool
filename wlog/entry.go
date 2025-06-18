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
	Skip(skip int) LoggerEntry
	Log()
}

type loggerEntry struct {
	level   zapcore.Level
	logger  *zap.Logger
	message string
	skip    int
}

func newLoggerEntry(level zapcore.Level, message string) LoggerEntry {
	return &loggerEntry{
		level:   level,
		logger:  logger,
		message: message,
		skip:    1, // 默认跳过1层调用者
	}
}

// 获取调用日志的函数或方法全名的最后一部分，一般来说是最后一个斜杠后的部分
// 对于函数，其全名为module/paths/pkg.FuncName(其中 paths 是从模块根目录到包的相对路径)，处理后返回的是pkg.FuncName
// 对于方法，其全名同理于函数，处理后返回的是pkg.(*Type).MethodName或者pkg.Type.MethodName
func callerName(skip int) string {
	pc, _, _, ok := runtime.Caller(skip + 1) // 这里需要+1，是因为要先跳转到调用callerName的Log方法
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

// 表示跳过调用栈中的若干层，用于控制日志中显示的调用位置。
// 例如：函数A调用函数B，函数B调用函数C，函数C中调用Error方法。
// 默认情况下跳过1层，日志将显示函数C中调用Error的代码位置；
// 若设置为跳过2层，则显示函数B中调用函数C的位置；
// 若设置为跳过3层，则显示函数A中调用函数B的位置。
func (l *loggerEntry) Skip(skip int) LoggerEntry {
	l.skip = skip
	return l
}

func (l *loggerEntry) Log() {
	l.logger.With(zap.String("caller", callerName(l.skip))).
		WithOptions(zap.AddCallerSkip(l.skip)).Check(l.level, l.message).Write()
}
