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
	config := loadConfig()
	var err error
	logger, err = config.Build()
	if err != nil {
		log.Fatalf("failed to initialize logger, err: %v", err)
	}
}

func loadConfig() zap.Config {
	config := zap.Config{}
	// 仅当明确是开发环境时，才使用Development配置，其余环境（测试、预发、生产等）统一使用结构化日志
	if os.Getenv("env") == "development" {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "time"
		config.EncoderConfig.MessageKey = "message"
		config.EncoderConfig.CallerKey = "line"
		config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	}
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	config.EncoderConfig.EncodeTime = customTimeEncoder
	config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	config.DisableStacktrace = true
	return config
}

func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	loc, err := time.LoadLocation("Asia/Shanghai") // 使用CST时间
	if err != nil {
		loc = time.UTC // 如果加载时区出错，则使用UTC时间
	}
	t = t.In(loc)
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}
