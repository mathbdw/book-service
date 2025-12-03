package repositories

import (
	"context"

	"github.com/mathbdw/book/internal/domain/entities"
)

//go:generate mockgen -destination=./../../../mocks/mock_book_repository.go -package=mocks -source=./book_repository.go

type BookRepository interface {
	Create(ctx context.Context, book entities.Book) (int64, error)
	GetByIDs(ctx context.Context, IDs []int64) ([]entities.Book, error)
	List(ctx context.Context, params entities.PaginationParams) (*entities.ResponseBooks, error)
	Remove(ctx context.Context, IDs []int64) error
}
