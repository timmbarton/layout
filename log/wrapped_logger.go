package log

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type WrappedLogger struct {
	l *zap.Logger
}

func NewWrappedLogger(l *zap.Logger) *WrappedLogger {
	if l == nil {
		l = zap.NewNop()
	}

	return &WrappedLogger{l: l}
}

func (w *WrappedLogger) Log(ctx context.Context, lvl zapcore.Level, msg string, fields ...zap.Field) {
	if ctx != nil {
		fields = append(fields, zap.Any(CtxKey, ctx))
	}

	w.l.Log(lvl, msg, fields...)
}
func (w *WrappedLogger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx != nil {
		fields = append(fields, zap.Any(CtxKey, ctx))
	}

	w.l.Debug(msg, fields...)
}
func (w *WrappedLogger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx != nil {
		fields = append(fields, zap.Any(CtxKey, ctx))
	}

	w.l.Info(msg, fields...)
}
func (w *WrappedLogger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx != nil {
		fields = append(fields, zap.Any(CtxKey, ctx))
	}

	w.l.Warn(msg, fields...)
}
func (w *WrappedLogger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx != nil {
		fields = append(fields, zap.Any(CtxKey, ctx))
	}

	w.l.Error(msg, fields...)
}
func (w *WrappedLogger) DPanic(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx != nil {
		fields = append(fields, zap.Any(CtxKey, ctx))
	}

	w.l.DPanic(msg, fields...)
}
func (w *WrappedLogger) Panic(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx != nil {
		fields = append(fields, zap.Any(CtxKey, ctx))
	}

	w.l.Panic(msg, fields...)
}
func (w *WrappedLogger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx != nil {
		fields = append(fields, zap.Any(CtxKey, ctx))
	}

	w.l.Fatal(msg, fields...)
}

func (w *WrappedLogger) Sync() error        { return w.l.Sync() }
func (w *WrappedLogger) Core() zapcore.Core { return w.l.Core() }
