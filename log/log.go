package log

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const CtxKey = "context"

func Named(s string) *WrappedLogger {
	return NewWrappedLogger(zap.L().Named(s))
}
func WithOptions(opts ...zap.Option) *WrappedLogger {
	return NewWrappedLogger(zap.L().WithOptions(opts...))
}
func With(fields ...zap.Field) *WrappedLogger {
	return NewWrappedLogger(zap.L().With(fields...))
}
func WithLazy(fields ...zap.Field) *WrappedLogger {
	return NewWrappedLogger(zap.L().WithLazy(fields...))
}

func Log(ctx context.Context, lvl zapcore.Level, msg string, fields ...zap.Field) {
	if ctx != nil {
		fields = append(fields, zap.Any(CtxKey, ctx))
	}

	zap.L().Log(lvl, msg, fields...)
}
func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx != nil {
		fields = append(fields, zap.Any(CtxKey, ctx))
	}

	zap.L().Debug(msg, fields...)
}
func Info(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx != nil {
		fields = append(fields, zap.Any(CtxKey, ctx))
	}

	zap.L().Info(msg, fields...)
}
func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx != nil {
		fields = append(fields, zap.Any(CtxKey, ctx))
	}

	zap.L().Warn(msg, fields...)
}
func Error(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx != nil {
		fields = append(fields, zap.Any(CtxKey, ctx))
	}

	zap.L().Error(msg, fields...)
}
func DPanic(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx != nil {
		fields = append(fields, zap.Any(CtxKey, ctx))
	}

	zap.L().DPanic(msg, fields...)
}
func Panic(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx != nil {
		fields = append(fields, zap.Any(CtxKey, ctx))
	}

	zap.L().Panic(msg, fields...)
}
func Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx != nil {
		fields = append(fields, zap.Any(CtxKey, ctx))
	}

	zap.L().Fatal(msg, fields...)
}

func Sync() error        { return zap.L().Sync() }
func Core() zapcore.Core { return zap.L().Core() }

func Json(key string, v any) zap.Field {
	data, _ := json.Marshal(v)
	return zap.String(key, string(data))
}
