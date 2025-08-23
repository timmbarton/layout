package signoz

import (
	"context"
	"errors"
	"fmt"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	ServiceName string `validate:"required"`
	URL         string `validate:"required"`
	Insecure    bool
}

type Connector struct {
	cfg Config

	res *resource.Resource

	log struct {
		exporter  *otlploghttp.Exporter
		processor log.Processor
		provider  *log.LoggerProvider
		logger    *zap.Logger
	}

	trace struct {
		exporter *otlptrace.Exporter
		provider *sdktrace.TracerProvider
	}
}

func New(cfg Config, loggerOpts ...zap.Option) (c *Connector, err error) {
	c = &Connector{
		cfg: cfg,
	}

	c.res, err = resource.New(
		nil,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(c.cfg.ServiceName),
			semconv.DeploymentEnvironmentKey.String(os.Getenv("ENV")),
			attribute.String("environment", os.Getenv("ENV")),
		),
	)
	if err != nil {
		return c, err
	}

	c.log.logger = new(zap.Logger)

	exporterOpts := []otlploghttp.Option{
		otlploghttp.WithEndpoint(c.cfg.URL),
	}

	if c.cfg.Insecure {
		exporterOpts = append(exporterOpts, otlploghttp.WithInsecure())
	}

	c.log.exporter, err = otlploghttp.New(
		nil,
		exporterOpts...,
	)
	if err != nil {
		return nil, err
	}

	c.log.processor = log.NewBatchProcessor(c.log.exporter)

	c.log.provider = log.NewLoggerProvider(
		log.WithProcessor(c.log.processor),
		log.WithResource(c.res),
	)

	loggerOpts = append(
		[]zap.Option{
			zap.AddCaller(),
			zap.AddStacktrace(zapcore.PanicLevel),
		},
		loggerOpts...,
	)

	c.log.logger = zap.New(
		otelzap.NewCore(
			"",
			otelzap.WithLoggerProvider(c.log.provider),
		),
		loggerOpts...,
	)

	zap.ReplaceGlobals(c.log.logger)

	return c, nil
}

func (c *Connector) Start(ctx context.Context) error {
	err := error(nil)

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(c.cfg.URL),
	}

	if c.cfg.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	c.trace.exporter, err = otlptrace.New(
		ctx,
		otlptracehttp.NewClient(opts...),
	)
	if err != nil {
		return err
	}

	c.trace.provider = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(c.trace.exporter),
		sdktrace.WithResource(c.res),
	)

	otel.SetTracerProvider(c.trace.provider)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	return nil
}

func (c *Connector) Stop(ctx context.Context) error {
	err := error(nil)

	if c.log.logger != nil {
		err = errors.Join(err, c.log.logger.Sync())
	}

	if c.log.provider != nil {
		err = errors.Join(err, c.log.provider.Shutdown(ctx))
	}

	if c.log.processor != nil {
		err = errors.Join(err, c.log.processor.Shutdown(ctx))
	}

	if c.log.exporter != nil {
		err = errors.Join(err, c.log.exporter.Shutdown(ctx))
	}

	if c.trace.provider != nil {
		err = errors.Join(err, c.trace.provider.Shutdown(ctx))
	}

	if c.trace.exporter != nil {
		err = errors.Join(err, c.trace.exporter.Shutdown(ctx))
	}

	return err
}

func (c *Connector) GetName() string {
	return fmt.Sprintf("SigNoz Logging and Tracing (%s)", c.cfg.ServiceName)
}

func (c *Connector) GetLogger() *zap.Logger {
	return c.log.logger
}
