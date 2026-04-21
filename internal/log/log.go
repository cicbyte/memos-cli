package log

import (
	"fmt"
	"time"

	"github.com/cicbyte/memos-cli/internal/common"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/gorm/logger"
)

var Logger *zap.Logger

type gormLoggerAdapter struct{}

func (g *gormLoggerAdapter) Printf(format string, v ...interface{}) {
	Logger.Info(fmt.Sprintf(format, v...))
}

func GetGormLogger() logger.Interface {
	return logger.New(
		&gormLoggerAdapter{},
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Warn,
			Colorful:      false,
		},
	)
}

func Init(logPath string) error {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 添加动态日志级别配置
	level := zap.NewAtomicLevel()
	switch common.GetAppConfig().Log.Level {
	case "debug":
		level.SetLevel(zapcore.DebugLevel)
	case "info":
		level.SetLevel(zapcore.InfoLevel)
	case "warn":
		level.SetLevel(zapcore.WarnLevel)
	case "error":
		level.SetLevel(zapcore.ErrorLevel)
	default:
		level.SetLevel(zapcore.InfoLevel)
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(&lumberjack.Logger{
			Filename:   logPath,
			MaxSize:    common.GetAppConfig().Log.MaxSize, // MB
			MaxBackups: common.GetAppConfig().Log.MaxBackups,
			MaxAge:     common.GetAppConfig().Log.MaxAge, // days
			Compress:   common.GetAppConfig().Log.Compress,
		}),
		level,
	)

	Logger = zap.New(core)
	return nil
}

func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}
