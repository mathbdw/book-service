package book

import (
	"context"

	"github.com/mathbdw/book/internal/domain/entities"
	"github.com/mathbdw/book/internal/errors"
	"github.com/mathbdw/book/internal/interfaces/observability"
	"github.com/mathbdw/book/internal/interfaces/repositories"
)

type ListBookUsecase struct {
	repoBook repositories.BookRepository
	observ   observability.UsecaseObservability
}

// NewListBookUsecase - Constructor AddBookUsecase
func NewListBookUsecase(repo repositories.BookRepository, observ observability.UsecaseObservability) ListBookUsecase {
	return ListBookUsecase{repoBook: repo, observ: observ}
}

// // Execute - Returns slice book by PaginationParams
func (uc *ListBookUsecase) Execute(ctx context.Context, params entities.PaginationParams) (*entities.ResponseBooks, error) {
	ctx, span := uc.observ.StartSpan(ctx, "ListBookUsecase")

	defer span.End()

	resp, err := uc.repoBook.List(ctx, params)
	if err != nil {
		span.SetAttributes([]observability.Attribute{{Key: "repo.book.failed", Value: true}})

		return nil, errors.Wrap(err, "listBookUsecase.Execute: list books")
	}

	return resp, nil
}
