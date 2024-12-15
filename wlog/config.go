package wlog

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"time"
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
	// 根据具体的环境变量进行调整
	if os.Getenv("env") == "production" {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "time"
		config.EncoderConfig.MessageKey = "message"
		config.EncoderConfig.CallerKey = "line"
		config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	config.EncoderConfig.EncodeTime = customTimeEncoder
	config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	config.DisableStacktrace = true
	return config
}

func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	loc, err := time.LoadLocation("Asia/Shanghai") // 使用 CST 时间
	if err != nil {
		loc = time.UTC // 如果加载时区出错，则使用 UTC 时间
	}
	t = t.In(loc)
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}
