package global

import (
	"context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"skis-admin-backend/config"
	"time"
)

var (
	Config *config.Config
	Lg     *zap.Logger
	DB     *gorm.DB
)

type ZapLogger struct {
	zapLogger *zap.Logger
}

func NewZapLogger(zapLogger *zap.Logger) *ZapLogger {
	return &ZapLogger{zapLogger: zapLogger}
}

func (z *ZapLogger) LogMode(level logger.LogLevel) logger.Interface {
	return z
}

func (z *ZapLogger) Log(lvl zapcore.Level, msg string, fields ...zap.Field) {
	z.zapLogger.Log(lvl, msg, fields...)
}

func (z *ZapLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	z.zapLogger.Info(msg, zap.Any("data", data))
}

func (z *ZapLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	z.zapLogger.Warn(msg, zap.Any("data", data))
}

func (z *ZapLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	z.zapLogger.Error(msg, zap.Any("data", data))
}

func (z *ZapLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	if err != nil {
		z.zapLogger.Error("SQL 执行错误",
			zap.Error(err),
			zap.String("sql", sql),
			zap.Int64("rows", rows),
			zap.Duration("elapsed", elapsed),
		)
	} else {
		z.zapLogger.Debug("SQL 执行成功",
			zap.String("sql", sql),
			zap.Int64("rows", rows),
			zap.Duration("elapsed", elapsed),
		)
	}
}
