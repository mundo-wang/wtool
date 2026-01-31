package wlog

import (
	"log"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func init() {
	zapConfig := loadZapConfig()
	var err error
	logger, err = zapConfig.Build()
	if err != nil {
		log.Fatalf("failed to initialize logger, err: %v", err)
	}
}

func loadZapConfig() zap.Config {
	zapConfig := zap.Config{}
	// 仅当明确是开发环境时，才使用Development配置，其余环境（测试、预发、生产等）统一使用结构化日志
	if os.Getenv("env") == "development" {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zapConfig.OutputPaths = []string{"stdout"}
	} else {
		zapConfig = zap.NewProductionConfig()
		zapConfig.EncoderConfig.TimeKey = "time"
		zapConfig.EncoderConfig.MessageKey = "message"
		zapConfig.EncoderConfig.CallerKey = "line"
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	}
	zapConfig.EncoderConfig.EncodeTime = customTimeEncoder
	zapConfig.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	zapConfig.DisableStacktrace = true
	return zapConfig
}

func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	loc, err := time.LoadLocation("Asia/Shanghai") // 使用CST时间
	if err != nil {
		loc = time.UTC // 如果加载时区出错，则使用UTC时间
	}
	t = t.In(loc)
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}
