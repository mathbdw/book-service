package metrics

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	
	"github.com/mathbdw/book/internal/interfaces/observability"
)

type opentelemetryRepositoryMetrics struct {
	meter metric.Meter

	// DB metrics
	dbQueryCounter  metric.Int64Counter
	dbQueryDuration metric.Float64Histogram
}

// NewOpentelemetryRepositoryMetrics - constructor opentelemetryRepositoryMetrics
func NewOpentelemetryRepositoryMetrics(mp *sdkmetric.MeterProvider) (observability.RepositoriesMetrics, error) {
	if mp == nil {
		return nil, fmt.Errorf("repositoryMetic.New: meter provider is required")
	}

	meter := mp.Meter("repository")

	dbQueryCounter, err := meter.Int64Counter(
		"database.queries.total",
		metric.WithDescription("Total number of create new book"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, fmt.Errorf("repositoryMetic.New: failed to create request counter: %w", err)
	}

	dbQueryDuration, err := meter.Float64Histogram(
		"database.query.duration.seconds",
		metric.WithDescription("DB request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("repositoryMetic.New: failed to create duration histogram: %w", err)
	}

	return &opentelemetryRepositoryMetrics{
		meter:           meter,
		dbQueryCounter:  dbQueryCounter,
		dbQueryDuration: dbQueryDuration,
	}, nil
}

// RecordDatabaseQuery - increments the counter and adds an entry
func (m *opentelemetryRepositoryMetrics) RecordDatabaseQuery(ctx context.Context, operation, table string, duration float64, success bool) {
	if operation == "" || table == "" {
		return
	}

	attributes := []attribute.KeyValue{
		attribute.String("operation", operation),
		attribute.String("table", table),
		attribute.Bool("success", success),
	}

	m.dbQueryCounter.Add(ctx, 1, metric.WithAttributes(attributes...))
	m.dbQueryDuration.Record(ctx, duration, metric.WithAttributes(attributes...))
}
