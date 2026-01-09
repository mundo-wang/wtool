package wlog

import (
	"context"
	"fmt"
	"path"
	"runtime"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type traceIdKeyType struct{}

var traceIdKey = traceIdKeyType{}

func WithTraceId(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIdKey, traceID)
}

type LoggerEntry interface {
	Ctx(ctx context.Context) LoggerEntry
	Field(key string, value interface{}) LoggerEntry
	Err(err error) LoggerEntry
	Skip(skip int) LoggerEntry

	Debug()
	Info()
	Warn()
	Error()
	Fatal()
	Panic()
}

type loggerEntry struct {
	logger  *zap.Logger
	message string
	skip    int
}

func Msg(message string) LoggerEntry {
	return &loggerEntry{
		logger:  logger,
		message: message,
		skip:    2, // 默认跳过2层调用者，write占1层，日志等级方法（如Error()）占1层
	}
}

func Msgf(format string, args ...interface{}) LoggerEntry {
	return Msg(fmt.Sprintf(format, args...))
}

// 获取调用日志的函数或方法全名的最后一部分，一般来说是最后一个斜杠后的部分
// 对于函数，其全名为module/paths/pkg.FuncName(其中paths是从模块根目录到包的相对路径)，处理后返回的是pkg.FuncName
// 对于方法，其全名同理于函数，处理后返回的是pkg.(*Type).MethodName或者pkg.Type.MethodName
func callerName(skip int) string {
	pc, _, _, ok := runtime.Caller(skip + 1) // 这里需要+1，是因为要先跳转到调用callerName的write方法
	if ok {
		fullName := runtime.FuncForPC(pc).Name()
		baseName := path.Base(fullName)
		return baseName
	} else {
		return "unknown"
	}
}

func (l *loggerEntry) Ctx(ctx context.Context) LoggerEntry {
	if ctx != nil {
		traceId, ok := ctx.Value(traceIdKey).(string)
		if ok && traceId != "" {
			l.logger = l.logger.With(zap.String("trace_id", traceId))
		}
	}
	return l
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
// 例如：函数A调用函数B，函数B调用函数C，函数C中打印日志。
// 若设置为跳过1层，将显示日志等级方法中调用write方法的位置；
// 默认的设置为跳过2层，将显示函数C中打印日志的位置；
// 若设置为跳过3层，则显示函数B中调用函数C的位置；
// 若设置为跳过4层，则显示函数A中调用函数B的位置。
func (l *loggerEntry) Skip(skip int) LoggerEntry {
	l.skip = skip
	return l
}

func (l *loggerEntry) write(level zapcore.Level) {
	l.logger.
		With(zap.String("caller", callerName(l.skip))).
		WithOptions(zap.AddCallerSkip(l.skip)).
		Check(level, l.message).
		Write()
}

func (l *loggerEntry) Debug() {
	l.write(zapcore.DebugLevel)
}

func (l *loggerEntry) Info() {
	l.write(zapcore.InfoLevel)
}

func (l *loggerEntry) Warn() {
	l.write(zapcore.WarnLevel)
}

func (l *loggerEntry) Error() {
	l.write(zapcore.ErrorLevel)
}

func (l *loggerEntry) Fatal() {
	l.write(zapcore.FatalLevel)
}

func (l *loggerEntry) Panic() {
	l.write(zapcore.PanicLevel)
}
