package observability

//go:generate mockgen -destination=./../../../mocks/mock_observability.go -package=mocks -source=./observability.go

// HandlerObservability - .
type HandlerObservability interface {
	Logger
	HandlersMetrics
	Tracer
}

// UsecaseObservability - .
type UsecaseObservability interface {
	// Logger
	UsecasesMetrics
	Tracer
}

// RepositoryObservability - .
type RepositoryObservability interface {
	// Logger
	RepositoriesMetrics
	Tracer
}
