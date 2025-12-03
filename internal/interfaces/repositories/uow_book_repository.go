package repositories

import (
	"context"
)

//go:generate mockgen -destination=./../../../mocks/mock_uow_book_repository.go -package=mocks -source=./uow_book_repository.go

type Repository struct {
	Book      BookRepository
	BookEvent BookEventRepository
}

type UnitOfWork interface {
	Do(ctx context.Context, fn func(repo *Repository) error) error
}
