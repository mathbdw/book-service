package book

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mathbdw/book/internal/domain/entities"
	"github.com/mathbdw/book/internal/errors"
	"github.com/mathbdw/book/internal/interfaces/observability"
	"github.com/mathbdw/book/internal/interfaces/repositories"
)

type AddBookUsecase struct {
	repoUOW repositories.UnitOfWork
	observ  observability.UsecaseObservability
}

// NewBookUsecase - Constructor AddBookUsecase
func NewAddBookUsecase(uow repositories.UnitOfWork, observ observability.UsecaseObservability) AddBookUsecase {
	return AddBookUsecase{repoUOW: uow, observ: observ}
}

// Add - Adds new book and book_event.
func (uc *AddBookUsecase) Execute(ctx context.Context, book entities.Book) error {
	start := time.Now()
	ctx, span := uc.observ.StartSpan(ctx, "AddBookUsecase")

	defer span.End()

	defer func() {
		duration := time.Since(start).Seconds()
		uc.observ.RecordBookCreated(ctx, book.Genre, duration)
	}()

	err := uc.repoUOW.Do(ctx, func(repo *repositories.Repository) error {
		id, err := repo.Book.Create(ctx, book)
		if err != nil {
			span.SetAttributes([]observability.Attribute{{Key: "repo.book.failed", Value: true}})

			return errors.Wrap(err, "addBookUsecases.Execute: failed to save book")
		}
		book.ID = id

		strBook, err := json.Marshal(book)
		if err != nil {
			span.RecordError(err)
			span.SetAttributes([]observability.Attribute{{Key: "json.marshal.failed", Value: true}})

			return errors.Wrap(err, "addBookUsecases.Execute: json marshal book")
		}

		event := entities.BookEvent{BookId: book.ID, Type: entities.Created, Status: entities.EventStatusNew, Payload: strBook}
		event.ID, err = repo.BookEvent.Create(ctx, event)
		if err != nil {
			span.SetAttributes([]observability.Attribute{{Key: "repo.bookEvent.failed", Value: true}})

			return errors.Wrap(err, "addBookUsecases.Execute: create book event")
		}

		return nil
	})

	return err
}
