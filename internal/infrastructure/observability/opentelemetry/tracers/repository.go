package tracers

import (
	"fmt"
	
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/mathbdw/book/internal/interfaces/observability"
)

// NewOpentelemetryRepositoryTracer - constructor tracer for usecases layer
func NewOpentelemetryRepositoryTracer(tp *sdktrace.TracerProvider) (observability.Tracer, error) {
	if tp == nil {
		return nil, fmt.Errorf("repositoryTracer.New: tracer provider is required")
	}

	tracer := tp.Tracer("repository")

	return &opentelemetryTracer{
		tracer: tracer,
	}, nil
}
