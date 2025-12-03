package opentelemetry

import (
	"context"
	"time"

	"github.com/mathbdw/book/config"
	"github.com/mathbdw/book/pkg/opentelemetry"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

type Metric struct {
	exporter *otlpmetricgrpc.Exporter
	resource *resource.Resource
}

// New - constructor Metric
func New(ctx context.Context, cfg *config.Config, opts ...otlpmetricgrpc.Option) (*Metric, error) {
	exporter, err := otlpmetricgrpc.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	res, err := opentelemetry.NewResource(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &Metric{exporter, res}, nil
}

// Start - .
func (t *Metric) Start(ctx context.Context, cfg *config.Config) (*sdkmetric.MeterProvider, error) {
	interval := cfg.Metric.ExporterInterval
	timeout := cfg.Metric.ExporterTimeout

	if interval == 0 {
		interval = 60 * time.Second
	}
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(t.resource),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(
			t.exporter,
			sdkmetric.WithInterval(interval),
			sdkmetric.WithTimeout(timeout),
		)),
	)

	return mp, nil
}
