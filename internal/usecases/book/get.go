package book

import (
	"context"

	"github.com/mathbdw/book/internal/domain/entities"
	"github.com/mathbdw/book/internal/errors"
	"github.com/mathbdw/book/internal/interfaces/observability"
	"github.com/mathbdw/book/internal/interfaces/repositories"
)

type GetBookUsecase struct {
	repoBook repositories.BookRepository
	observ   observability.UsecaseObservability
}

// NewGetBookUsecase - Constructor GetBookUsecase
func NewGetBookUsecase(repo repositories.BookRepository, observ observability.UsecaseObservability) GetBookUsecase {
	return GetBookUsecase{repoBook: repo, observ: observ}
}

// GetByIDs - Returns slice book by IDs
func (uc *GetBookUsecase) GetByIDs(ctx context.Context, IDs []int64) ([]entities.Book, error) {
	ctx, span := uc.observ.StartSpan(ctx, "GetBookUsecase")

	defer span.End()

	books, err := uc.repoBook.GetByIDs(ctx, IDs)
	if err != nil {
		span.SetAttributes([]observability.Attribute{{Key: "repo.book.failed", Value: true}})

		return []entities.Book{}, errors.Wrap(err, "GetBookUsecase.GetByIds: get books")
	}

	return books, nil
}
