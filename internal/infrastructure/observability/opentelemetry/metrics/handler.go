package metrics

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/mathbdw/book/internal/interfaces/observability"
)

type opentelemetryHandlerMetrics struct {
	meter metric.Meter

	// Handler Grps metrics
	handlerCounter  metric.Int64Counter
	handlerDuration metric.Float64Histogram
}

// NewOpentelemetryHandlerMetrics - constructor opentelemetryHandlerMetrics
func NewOpentelemetryHandlerMetrics(mp *sdkmetric.MeterProvider) (observability.HandlersMetrics, error) {
	if mp == nil {
		return nil, fmt.Errorf("handlerMetic.New: meter provider is required")
	}

	meter := mp.Meter(
		"handler",
		metric.WithInstrumentationVersion("v1"),
	)

	handlerCounter, err := meter.Int64Counter(
		"http.requests.total",
		metric.WithDescription("Total number of GRPC requests"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, fmt.Errorf("handlerMetic.New: failed to create request counter: %w", err)
	}

	handlerDuration, err := meter.Float64Histogram(
		"http.request.duration.seconds",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("metic.New: failed to create duration histogram: %w", err)
	}

	return &opentelemetryHandlerMetrics{
		meter:           meter,
		handlerCounter:  handlerCounter,
		handlerDuration: handlerDuration,
	}, nil
}

// RecordHanderRequest - increments the counter and adds an entry
func (m *opentelemetryHandlerMetrics) RecordHanderRequest(ctx context.Context, method, path string, statusCode int, duration float64) {
	if method == "" || path == "" {
		return
	}

	attributes := []attribute.KeyValue{
		attribute.String("method", method),
		attribute.String("path", path),
		attribute.Int("status", statusCode),
	}

	m.handlerCounter.Add(ctx, 1, metric.WithAttributes(attributes...))
	m.handlerDuration.Record(ctx, duration, metric.WithAttributes(attributes...))
}
