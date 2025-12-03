package publisher

import (
	"context"

	"github.com/mathbdw/book/internal/domain/entities"
)

//go:generate mockgen -destination=./../../../mocks/mock_event_publisher.go -package=mocks -source=./event_publisher.go

// EventPublisher - .
type EventPublisher interface {
	Publish(ctx context.Context, bookEvent *entities.BookEvent) error
}
