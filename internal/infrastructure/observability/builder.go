package observability

import (
	"fmt"

	"github.com/mathbdw/book/internal/interfaces/observability"
)

func NewObservabilityBuilder() *Observability {
	return &Observability{}
}

// WithHandlerMertic - add logger
func (b *Observability) WithLogger(log observability.Logger) *Observability {
	b.logger = log

	return b
}

// WithHandlerMertic - add metric for handlers
func (b *Observability) WithHandlerMertic(m observability.HandlersMetrics) *Observability {
	b.handlersMetrics = m

	return b
}

// WithUsecaseMertic - add metric for usecases
func (b *Observability) WithUsecaseMertic(m observability.UsecasesMetrics) *Observability {
	b.usecasesMetrics = m

	return b
}

// WithHandlerMertic - add metric for repositories
func (b *Observability) WithRepositoryMertic(m observability.RepositoriesMetrics) *Observability {
	b.repositoriesMetrics = m

	return b
}

// WithHandlerMertic - add tracer for handlers
func (b *Observability) WithHandlerTracer(t observability.Tracer) *Observability {
	b.handlersTracer = t

	return b
}

// WithUsecaseTracer - add tracer for usecases
func (b *Observability) WithUsecaseTracer(t observability.Tracer) *Observability {
	b.usecasesTracer = t

	return b
}

// WithHandlerMertic - add tracer for repositories
func (b *Observability) WithRepositoryTracer(t observability.Tracer) *Observability {
	b.repositoriesTracer = t

	return b
}

// Build - .
func (b *Observability) Build() (*Observability, error) {
	if b.logger == nil {
		return nil, fmt.Errorf("builder.Build: logger is required")
	}

	return &Observability{
		logger:              b.logger,
		handlersMetrics:     b.handlersMetrics,
		usecasesMetrics:     b.usecasesMetrics,
		repositoriesMetrics: b.repositoriesMetrics,
		handlersTracer:      b.handlersTracer,
		usecasesTracer:      b.usecasesTracer,
		repositoriesTracer:  b.repositoriesTracer,
	}, nil
}
