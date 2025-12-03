package repositories

import (
	"context"

	"github.com/mathbdw/book/internal/domain/entities"
)

//go:generate mockgen -destination=./../../../mocks/mock_book_event_repository.go -package=mocks -source=./book_event_repository.go



type BookEventRepository interface {
	Create(ctx context.Context, bookEvent entities.BookEvent) (int64, error)
	Lock(ctx context.Context, batchSize uint64) ([]entities.BookEvent, error)
	Unlock(ctx context.Context, eventIDs []int64) error
	Remove(ctx context.Context, eventIDs []int64) error
}
