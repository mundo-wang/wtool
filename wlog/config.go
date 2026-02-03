package wlog

import (
	"log"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger      *zap.Logger
	shanghaiLoc *time.Location
)

func init() {
	var err error
	shanghaiLoc, err = time.LoadLocation("Asia/Shanghai") // 使用CST时间
	if err != nil {
		shanghaiLoc = time.UTC // 如果加载时区出错，则使用UTC时间
		err = nil
	}
	zapConfig := loadZapConfig()
	logger, err = zapConfig.Build()
	if err != nil {
		log.Fatalf("failed to initialize logger, err: %v", err)
	}
}

func isDevEnv() bool {
	switch os.Getenv("ENV") {
	case "dev", "development", "local":
		return true
	default:
		return false
	}
}

func loadZapConfig() zap.Config {
	zapConfig := zap.Config{}
	// 仅当明确是开发环境时，才使用Development配置，其余环境（测试、预发、生产等）统一使用结构化日志
	if isDevEnv() {
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
	t = t.In(shanghaiLoc)
	enc.AppendString(t.Format(time.DateTime))
}
