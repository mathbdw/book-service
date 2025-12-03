package tracers

import (
	"fmt"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/mathbdw/book/internal/interfaces/observability"
)

// NewOpentelemetryUsecaseTracer - constructor tracer for usecases layer
func NewOpentelemetryUsecaseTracer(tp *sdktrace.TracerProvider) (observability.Tracer, error) {
	if tp == nil {
		return nil, fmt.Errorf("usecaseTracer.New: tracer provider is required")
	}

	tracer := tp.Tracer("usecase")

	return &opentelemetryTracer{
		tracer: tracer,
	}, nil
}
