package signoz

import (
	"context"
	"fmt"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	ServiceName  string `validate:"required"`
	OtlpEndpoint string `validate:"required"`
}

type Connector struct {
	cfg      Config
	logger   *zap.Logger
	provider *log.LoggerProvider
}

func New(cfg Config, opts ...zap.Option) (c *Connector, err error) {
	c = &Connector{
		cfg:    cfg,
		logger: new(zap.Logger),
	}

	exporter, err := otlploghttp.New(
		nil,
		otlploghttp.WithEndpoint(c.cfg.OtlpEndpoint),
		otlploghttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	// Create a log record processor pipeline.
	processor := log.NewBatchProcessor(exporter)

	// Create a logger provider.
	// You can pass this instance directly when creating a log bridge.
	c.provider = log.NewLoggerProvider(
		log.WithProcessor(processor),
	)

	// Initialize a zap logger with the otelzap bridge core.
	opts = append(
		[]zap.Option{
			zap.AddCaller(),
			zap.AddStacktrace(zapcore.PanicLevel),
		},
		opts...,
	)

	c.logger = zap.New(
		otelzap.NewCore(
			c.cfg.ServiceName,
			otelzap.WithLoggerProvider(c.provider),
		),
		opts...,
	)

	zap.ReplaceGlobals(c.logger)

	return c, nil
}

func (c *Connector) Start(_ context.Context) error {
	return nil
}

func (c *Connector) Stop(ctx context.Context) (err error) {
	if c.logger != nil {
		err = c.logger.Sync()
		if err != nil {
			return err
		}
	}

	if c.provider != nil {
		err = c.provider.Shutdown(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Connector) GetName() string { return fmt.Sprintf("SigNoz Logger (%s)", c.cfg.ServiceName) }

func (c *Connector) GetLogger() *zap.Logger {
	return c.logger
}
