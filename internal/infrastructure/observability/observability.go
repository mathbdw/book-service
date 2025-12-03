package observability

import "github.com/mathbdw/book/internal/interfaces/observability"

type Observability struct {
	logger              observability.Logger
	handlersMetrics     observability.HandlersMetrics
	usecasesMetrics     observability.UsecasesMetrics
	repositoriesMetrics observability.RepositoriesMetrics
	handlersTracer      observability.Tracer
	usecasesTracer      observability.Tracer
	repositoriesTracer  observability.Tracer
}

// handlers
type handlerObservability struct {
	observability.Logger
	observability.HandlersMetrics
	observability.Tracer
}

func (o *Observability) ForHandler() observability.HandlerObservability {
	return &handlerObservability{
		Logger:          o.logger,
		HandlersMetrics: o.handlersMetrics,
		Tracer:          o.handlersTracer,
	}
}

// usecases
type usecasesObservability struct {
	// observability.Logger
	observability.UsecasesMetrics
	observability.Tracer
}

func (o *Observability) ForUsecases() observability.UsecaseObservability {
	return &usecasesObservability{
		// Logger:          o.logger,
		UsecasesMetrics: o.usecasesMetrics,
		Tracer:          o.usecasesTracer,
	}
}

// repository
type repositoryObservability struct {
	observability.Logger
	observability.RepositoriesMetrics
	observability.Tracer
}

func (o *Observability) ForRepository() observability.RepositoryObservability {
	return &repositoryObservability{
		Logger:              o.logger,
		RepositoriesMetrics: o.repositoriesMetrics,
		Tracer:              o.repositoriesTracer,
	}
}
