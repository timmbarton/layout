package tracingconn

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

type Config struct {
	URL         string `validate:"required"`
	ServiceName string `validate:"required"`
}

type Tracer struct {
	cfg Config
	exp *jaeger.Exporter
	tp  *tracesdk.TracerProvider
}

//goland:noinspection ALL
func New(cfg Config) *Tracer {
	return &Tracer{
		cfg: cfg,
	}
}

func (t *Tracer) Start(_ context.Context) error {
	err := error(nil)

	t.exp, err = jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(t.cfg.URL)))
	if err != nil {
		return err
	}

	t.tp = tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(t.exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(t.cfg.ServiceName),
		)),
	)
	otel.SetTracerProvider(t.tp)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	return nil
}
func (t *Tracer) Stop(ctx context.Context) error {
	err1 := t.tp.Shutdown(ctx)
	err2 := t.exp.Shutdown(ctx)

	return errors.Join(err1, err2)
}
func (t *Tracer) GetName() string { return "Tracing" }
