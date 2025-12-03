package metrics

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	
	"github.com/mathbdw/book/internal/interfaces/observability"
)

type opentelemetryUsecaseMetrics struct {
	meter metric.Meter

	// usecase metrics
	usecaseCounter  metric.Int64Counter
	usecaseDuration metric.Float64Histogram
}

// NewOpentelemetryUsecaseMetrics - constructor opentelemetryUsecaseMetrics
func NewOpentelemetryUsecaseMetrics(mp *sdkmetric.MeterProvider) (observability.UsecasesMetrics, error) {
	if mp == nil {
		return nil, fmt.Errorf("usecaseMetic.New: meter provider is required")
	}

	meter := mp.Meter("usecase")

	usecaseCounter, err := meter.Int64Counter(
		"usecases.genre.total",
		metric.WithDescription("Total number of create new book"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, fmt.Errorf("usecaseMetic.New: failed to create request counter: %w", err)
	}

	usecaseDuration, err := meter.Float64Histogram(
		"usecases.genre.duration.seconds",
		metric.WithDescription("DB request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("usecaseMetic.New: failed to create duration histogram: %w", err)
	}

	return &opentelemetryUsecaseMetrics{
		meter:           meter,
		usecaseCounter:  usecaseCounter,
		usecaseDuration: usecaseDuration,
	}, nil
}

// RecordBookCreated - increments the counter and adds an entry
func (m *opentelemetryUsecaseMetrics) RecordBookCreated(ctx context.Context, genre string, duration float64) {
	if genre == "" {
		return
	}

	attributes := []attribute.KeyValue{
		attribute.String("book.genre", genre),
	}

	m.usecaseCounter.Add(ctx, 1, metric.WithAttributes(attributes...))
	m.usecaseDuration.Record(ctx, duration, metric.WithAttributes(attributes...))
}
