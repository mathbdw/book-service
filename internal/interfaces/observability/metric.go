package observability

import "context"

//go:generate mockgen -destination=./../../../mocks/mock_metric.go -package=mocks -source=./metric.go

type HandlersMetrics interface {
	RecordHanderRequest(ctx context.Context, method, path string, statusCode int, duration float64)
}

type UsecasesMetrics interface {
	RecordBookCreated(ctx context.Context, genre string, duration float64)
}

type RepositoriesMetrics interface {
	RecordDatabaseQuery(ctx context.Context, operation, table string, duration float64, success bool)
}
