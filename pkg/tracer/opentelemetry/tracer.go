package opentelemetry

import (
	"context"

	"github.com/mathbdw/book/config"
	"github.com/mathbdw/book/pkg/opentelemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Tracer struct {
	exporter *otlptrace.Exporter
	resource *resource.Resource
}

// New - constructor Tracer
func New(ctx context.Context, cfg *config.Config, opts ...otlptracegrpc.Option) (*Tracer, error) {
	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	res, err := opentelemetry.NewResource(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &Tracer{exporter, res}, nil
}

// Start - .
func (t *Tracer) Start(ctx context.Context, cfg *config.Config) (*sdktrace.TracerProvider, error) {
	var sampler sdktrace.Sampler

	if cfg.Project.Environment == "development" {
		sampler = sdktrace.AlwaysSample()
	} else {
		sampler = sdktrace.TraceIDRatioBased(0.01)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(t.exporter),
		sdktrace.WithResource(t.resource),
		sdktrace.WithSampler(sampler),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	return tp, nil
}