package services

import (
	"context"
	"errors"
	"sync"
	"time"

	errs "github.com/mathbdw/book/internal/errors"
	"github.com/mathbdw/book/internal/interfaces/observability"
	"github.com/mathbdw/book/internal/interfaces/publisher"
	"github.com/mathbdw/book/internal/interfaces/repositories"
)

type OutboxProcessor struct {
	eventRepo repositories.BookEventRepository
	publisher publisher.EventPublisher
	logger    observability.Logger

	batchSize    uint64
	interval     time.Duration
	countWorkers uint8
}

// New - constructor outbox processor
func New(repo repositories.BookEventRepository, publisher publisher.EventPublisher, logger observability.Logger, opts ...Option) *OutboxProcessor {
	op := &OutboxProcessor{
		eventRepo: repo,
		publisher: publisher,
		logger:    logger,
	}

	for _, opt := range opts {
		opt(op)
	}
	return op
}

// Start - sets semafor for outbox
func (op *OutboxProcessor) Start(ctx context.Context) error {
	var wg sync.WaitGroup
	errCh := make(chan error, op.countWorkers)
	defer close(errCh)

	ticker := time.NewTicker(op.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			op.logger.Info("outbox.Start: graceful shutdown", nil)
			op.logger.Info("outbox.Start: graceful shutdown",
				map[string]any{
					"reason":  ctx.Err().Error(),
					"workers": op.countWorkers,
				})

			return nil
		case err := <-errCh:
			wg.Wait()

			op.logger.Error("outbox.Start: graceful shutdown: critical error shutdown",
				map[string]any{
					"error":   err.Error(),
					"workers": op.countWorkers,
					"reason":  "worker_error",
				})

			return err
		case <-ticker.C:
			for i := uint8(0); i < op.countWorkers; i++ {
				wg.Add(1)
				go func(workNumber uint8) {
					defer wg.Done()
					if err := op.processEvent(ctx, workNumber); err != nil {
						errCh <- err
					}
				}(i)
			}
		}
	}
}

// processEvent - preparing event for sending
func (op *OutboxProcessor) processEvent(ctx context.Context, number uint8) error {
	op.logger.Debug("outbox.processEvent: run workers", map[string]any{"worker": number})

	events, err := op.eventRepo.Lock(ctx, op.batchSize)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			op.logger.Debug(
				"outbox.processEvent: no events found",
				map[string]any{"worker": number},
			)
			return nil
		}
		op.logger.Error(
			"outbox.processEvent: lock failed",
			map[string]any{"worker": number, "error": err},
		)

		return err
	}

	eventsIDsSuccess := make([]int64, 0, len(events))
	var errSend error

	defer func() {
		if len(eventsIDsSuccess) > 0 {
			err = op.eventRepo.Remove(ctx, eventsIDsSuccess)

			if err != nil {
				op.logger.Error(
					"outbox.processEvent: remove failed",
					map[string]any{"worker": number, "error": err, "eventIDs": eventsIDsSuccess},
				)
			}
		}
	}()

	for _, event := range events {
		err = op.publisher.Publish(ctx, &event)

		if err != nil {
			op.logger.Error(
				"outbox.processEvent: publish failed",
				map[string]any{"worker": number, "error": err, "eventID": event.ID},
			)

			errSend = err
			break
		}

		eventsIDsSuccess = append(eventsIDsSuccess, event.ID)
	}

	if errSend != nil {
		eventsIDsFailure := make([]int64, 0, len(events)-len(eventsIDsSuccess))

		for i := len(eventsIDsSuccess); i < len(events); i++ {
			eventsIDsFailure = append(eventsIDsFailure, events[i].ID)
		}

		err = op.eventRepo.Unlock(ctx, eventsIDsFailure)
		if err != nil {
			op.logger.Error(
				"outbox.processEvent: unlock failed",
				map[string]any{"worker": number, "error": err, "eventIDs": eventsIDsFailure},
			)

			return err
		}

		return errSend
	}

	return nil
}
