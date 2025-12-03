package bookevent

import (
	"context"

	"github.com/mathbdw/book/internal/domain/entities"
	errs "github.com/mathbdw/book/internal/errors"
	"github.com/mathbdw/book/internal/interfaces/observability"
	"github.com/mathbdw/book/internal/interfaces/repositories"
)

type BookEventUsecaseInterface interface {
	Lock(ctx context.Context, eventIds []int64)
	Unlock(ctx context.Context, eventIDs []int64)
	Remove(ctx context.Context, eventIDs []int64)
}

type BookEventUsecase struct {
	repo   repositories.BookEventRepository
	observ observability.UsecaseObservability
}

// NewBookEventUsecase - Constructor book event
func NewBookEventUsecase(repo repositories.BookEventRepository, observ observability.UsecaseObservability) BookEventUsecase {
	return BookEventUsecase{repo: repo, observ: observ}
}

// Lock - locks book events by the specified identifiers
// Returns a list of locked events or an error
func (uc *BookEventUsecase) Lock(ctx context.Context, batchSize uint64) ([]entities.BookEvent, error) {
	ctx, span := uc.observ.StartSpan(ctx, "BookEventUsecase.lock")

	defer span.End()

	span.SetAttributes([]observability.Attribute{{Key: "batchSize", Value: batchSize}})

	events, err := uc.repo.Lock(ctx, batchSize)
	if err != nil {
		span.SetAttributes([]observability.Attribute{{Key: "repo.BookEvent.failed", Value: true}})

		return []entities.BookEvent{}, errs.Wrap(err, "bookEventUsecase.Lock: set lock events")
	}

	return events, nil
}

// Unlock - unlocks book events for the specified IDs
// Returns an error if the execution fails
func (uc *BookEventUsecase) Unlock(ctx context.Context, eventIDs []int64) error {
	ctx, span := uc.observ.StartSpan(ctx, "BookEventUsecase.unlock")

	defer span.End()

	span.SetAttributes([]observability.Attribute{{Key: "eventIds", Value: eventIDs}})

	err := uc.repo.Unlock(ctx, eventIDs)
	if err != nil {
		span.SetAttributes([]observability.Attribute{{Key: "repo.BookEvent.failed", Value: true}})

		return errs.Wrap(err, "bookEventUsecase.Unlock: set unlock events")
	}

	return nil
}

// Remove - deletes rows book events for the specified IDs
// Returns an error if the execution fails
func (uc *BookEventUsecase) Remove(ctx context.Context, eventIDs []int64) error {
	ctx, span := uc.observ.StartSpan(ctx, "BookEventUsecase.remove")

	defer span.End()

	span.SetAttributes([]observability.Attribute{{Key: "eventIds", Value: eventIDs}})

	err := uc.repo.Remove(ctx, eventIDs)
	if err != nil {
		span.SetAttributes([]observability.Attribute{{Key: "repo.BookEvent.failed", Value: true}})

		return errs.Wrap(err, "bookEventUsecase.Remove: delete rows events")
	}

	return nil
}
