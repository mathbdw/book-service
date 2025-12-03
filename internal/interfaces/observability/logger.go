package observability

import "context"

type Field map[string]any

//go:generate mockgen -destination=./../../../mocks/mock_logger.go -package=mocks -source=./logger.go

type Logger interface {
	WithContext(ctx context.Context) Logger
	SetLevel(newLevel int8)
	Debug(msg string, fields Field)
	Info(msg string, fields Field)
	Warn(msg string, fields Field)
	Error(msg string, fields Field)
	Fatal(msg string, fields Field)
}
