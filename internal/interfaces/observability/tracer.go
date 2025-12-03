package observability

import "context"

type Attribute struct {
	Key   string
	Value any
}

//go:generate mockgen -destination=./../../../mocks/mock_tracer.go -package=mocks -source=./tracer.go

type Tracer interface {
	StartSpan(ctx context.Context, name string) (context.Context, Span)
}

type Span interface {
	End()
	RecordError(err error)
	SetAttributes(attrs []Attribute)
}
