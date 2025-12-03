package tracers

import (
	"fmt"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	
	"github.com/mathbdw/book/internal/interfaces/observability"
)

// NewOpentelemetryHandlerTracer - constructor tracer for handler
func NewOpentelemetryHandlerTracer(tp *sdktrace.TracerProvider) (observability.Tracer, error) {
	if tp == nil {
		return nil, fmt.Errorf("handlerTracer.New: tracer provider is required")
	}

	tracer := tp.Tracer("handler")

	return &opentelemetryTracer{
		tracer: tracer,
	}, nil
}
