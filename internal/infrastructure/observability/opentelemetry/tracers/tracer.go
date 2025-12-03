package tracers

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	
	"github.com/mathbdw/book/internal/interfaces/observability"
)

type opentelemetryTracer struct {
	tracer trace.Tracer
}

// Start - creates a span and a context.Context containing the newly-created span.
func (t *opentelemetryTracer) StartSpan(ctx context.Context, name string) (context.Context, observability.Span) {
	ctx, span := t.tracer.Start(ctx, name)
	return ctx, &opentelemetrySpan{span: span}
}

type opentelemetrySpan struct {
	span trace.Span
}

// End - completes the Span.
func (s *opentelemetrySpan) End() {
	s.span.End()
}

// RecordError - records an error along with a stack trace
func (s *opentelemetrySpan) RecordError(err error) {
	s.span.RecordError(err)
}

// SetAttributes - sets kv as attributes of the Span
func (s *opentelemetrySpan) SetAttributes(attrs []observability.Attribute) {
	openAttr := make([]attribute.KeyValue, 0, len(attrs))

	for _, attr := range attrs {
		openAttr = append(openAttr, convertAttributtes(attr))
	}

	s.span.SetAttributes(openAttr...)
}

// convertAttributtes - convertes attributes to type Opentelemetry attribute.KeyValue
func convertAttributtes(attr observability.Attribute) attribute.KeyValue {
	switch v := attr.Value.(type) {
	case bool:
		return attribute.Bool(attr.Key, v)
	case []bool:
		return attribute.BoolSlice(attr.Key, v)
	case float64:
		return attribute.Float64(attr.Key, v)
	case []float64:
		return attribute.Float64Slice(attr.Key, v)
	case int:
		return attribute.Int(attr.Key, v)
	case []int:
		return attribute.IntSlice(attr.Key, v)
	case string:
		return attribute.String(attr.Key, v)
	case []string:
		return attribute.StringSlice(attr.Key, v)
	default:
		return attribute.String(attr.Key, fmt.Sprintf("%+v", v))
	}
}
