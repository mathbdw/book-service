package book

import (
	"context"
	"fmt"

	"github.com/mathbdw/book/internal/domain/entities"
	"github.com/mathbdw/book/internal/errors"
	"github.com/mathbdw/book/internal/interfaces/observability"
	"github.com/mathbdw/book/internal/interfaces/repositories"
)

type RemoveBookUsecase struct {
	repoUOW repositories.UnitOfWork
	observ  observability.UsecaseObservability
}

// NewRemoveBookUsecase - Constructor RemoveBookUsecase
func NewRemoveBookUsecase(uow repositories.UnitOfWork, observ observability.UsecaseObservability) RemoveBookUsecase {
	return RemoveBookUsecase{repoUOW: uow, observ: observ}
}

// Remove - Update field removed of Book and create rows book_event.
func (uc *RemoveBookUsecase) Execute(ctx context.Context, IDs []int64) error {
	ctx, span := uc.observ.StartSpan(ctx, "RemoveBookUsecase")

	defer span.End()

	err := uc.repoUOW.Do(ctx, func(repo *repositories.Repository) error {
		err := repo.Book.Remove(ctx, IDs)
		if err != nil {
			span.SetAttributes([]observability.Attribute{{Key: "repo.book.failed", Value: true}})

			return errors.Wrap(err, "RemoveBookUsecase.Execute: remove Book")
		}

		for _, id := range IDs {
			event := entities.BookEvent{BookId: int64(id), Type: entities.Deleted, Status: entities.EventStatusNew}
			event.ID, err = repo.BookEvent.Create(ctx, event)
			if err != nil {
				span.SetAttributes([]observability.Attribute{{Key: "repo.bookEvent.failed", Value: true}})

				return errors.Wrap(err, fmt.Sprintf("RemoveBookUsecase.Execute: create Book Event = %d", id))
			}
		}

		return nil
	})

	return err
}
