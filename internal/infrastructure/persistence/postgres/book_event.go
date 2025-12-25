package postgres

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/mathbdw/book/internal/domain/entities"
	"github.com/mathbdw/book/internal/errors"
	"github.com/mathbdw/book/internal/interfaces/observability"
	"github.com/mathbdw/book/internal/interfaces/repositories"
)

type bookEventRepository struct {
	querier sqlx.ExtContext
	builder sq.StatementBuilderType

	observ observability.RepositoryObservability
}

// NewBookEventRepository - Constructor BookEventRepository
func NewBookEventRepository(querier sqlx.ExtContext, builder sq.StatementBuilderType, observ observability.RepositoryObservability) repositories.BookEventRepository {
	return &bookEventRepository{querier: querier, builder: builder, observ: observ}
}

func (r *bookEventRepository) Create(ctx context.Context, bookEvent entities.BookEvent) (int64, error) {
	var success bool
	start := time.Now()
	ctx, span := r.observ.StartSpan(ctx, "bookEventRepository.create")
	span.SetAttributes([]observability.Attribute{
		{Key: "bookEvent.bookId", Value: bookEvent.BookId},
		{Key: "bookEvent.type", Value: int(bookEvent.Type)},
		{Key: "bookEvent.status", Value: int(bookEvent.Status)},
		{Key: "bookEvent.payload", Value: string(bookEvent.Payload)},
	})
	defer span.End()

	defer func() {
		duration := time.Since(start).Seconds()
		r.observ.RecordDatabaseQuery(ctx, "insert", "book_event", duration, success)
	}()

	data := map[string]interface{}{
		"book_id": bookEvent.BookId,
		"type":    bookEvent.Type,
		"status":  bookEvent.Status,
		"payload": bookEvent.Payload,
	}
	query, args, err := r.builder.Insert("book_event").SetMap(data).Suffix("RETURNING id").ToSql()
	if err != nil {
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "toSql.failed", Value: true}})

		return 0, errors.Wrap(err, "bookEventPostgres.Create: building query")
	}

	var id int64
	err = r.querier.QueryRowxContext(ctx, query, args...).Scan(&id)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "scan.failed", Value: true}})

		return 0, errors.Wrap(err, "bookEventPostgres.Create: scanning query")
	}

	success = true
	return id, nil
}

// Lock - Sets status lock
func (r *bookEventRepository) Lock(ctx context.Context, batchSize uint64) ([]entities.BookEvent, error) {
	var success bool
	start := time.Now()
	ctx, span := r.observ.StartSpan(ctx, "bookEventRepository.look")
	span.SetAttributes([]observability.Attribute{{Key: "batchSize", Value: batchSize}})

	defer span.End()

	defer func() {
		duration := time.Since(start).Seconds()
		r.observ.RecordDatabaseQuery(ctx, "update", "book_event", duration, success)
	}()

	// 1. SELECT с блокировкой
	lockQuery := r.builder.Select("id").
		From("book_event").
		Where(sq.NotEq{"status": entities.EventStatusLock}).
		OrderBy("id ASC").
		Limit(batchSize).
		Suffix("FOR UPDATE SKIP LOCKED")

	lockSQL, lockArgs, err := lockQuery.ToSql()

	if err != nil {
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "toSql.failed", Value: true}})

		return []entities.BookEvent{}, errors.Wrap(err, "bookEventPostgres.Lock: building query")
	}

	// 2. Ручной CTE запрос
	query := `
        WITH locked_event AS (` + lockSQL + `)
        UPDATE book_event 
        SET status = $2
        WHERE id IN (SELECT id FROM locked_event)
        RETURNING id, book_id, type, payload
    `
	args := append([]any{entities.EventStatusLock}, lockArgs...)
	rows, err := r.querier.QueryxContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "queryxContext.failed", Value: true}})

		return nil, errors.Wrap(err, "bookEventPostgres.Lock: executing query")
	}
	defer rows.Close()

	events := make([]entities.BookEvent, 0, batchSize)
	for rows.Next() {
		var event entities.BookEvent
		err = rows.StructScan(&event)
		if err != nil {
			span.RecordError(err)
			span.SetAttributes([]observability.Attribute{{Key: "scan.failed", Value: true}})

			return nil, errors.Wrap(err, "bookEventPostgres.Lock: scanning row")
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "iteration.failed", Value: true}})

		return nil, errors.Wrap(err, "bookEventPostgres.Lock: iteration rows")
	}

	success = true
	if len(events) == 0 {
		span.SetAttributes([]observability.Attribute{{Key: "len.bookEvent.zero.failed", Value: true}})

		return events, errors.ErrNotFound
	}

	return events, nil
}

// Unlock - Sets the unlock status for locked rows.
func (r *bookEventRepository) Unlock(ctx context.Context, eventIDs []int64) error {
	var success bool
	start := time.Now()
	ctx, span := r.observ.StartSpan(ctx, "bookEventRepository.unlook")
	span.SetAttributes([]observability.Attribute{{Key: "eventIDs", Value: eventIDs}})

	defer span.End()

	defer func() {
		duration := time.Since(start).Seconds()
		r.observ.RecordDatabaseQuery(ctx, "update", "book_event", duration, success)
	}()

	query, args, err := r.builder.Update("book_event").
		Set("status", entities.EventStatusUnlock).
		Where(sq.And{sq.Eq{"id": eventIDs}, sq.Eq{"status": entities.EventStatusLock}}).
		ToSql()

	if err != nil {
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "toSql.failed", Value: true}})

		return errors.Wrap(err, "bookEventPostgres.Unlock: building query")
	}

	res, err := r.querier.ExecContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "scan.failed", Value: true}})

		return errors.Wrap(err, "bookEventPostgres.Unlock: executing query")
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "rowsAffected.failed", Value: true}})

		return errors.Wrap(err, "bookEventPostgres.Unlock: getting rows affected")
	}

	if rowsAffected != int64(len(eventIDs)) {
		span.SetAttributes([]observability.Attribute{{Key: "len.bookEvent.noEqual.failed", Value: true}})

		return errors.New(fmt.Sprintf("bookEventPostgres.Unlock: expected rowsAffected %d, actual %d", len(eventIDs), rowsAffected))
	}

	success = true
	return nil
}

// Remove - Removes locked rows.
func (r *bookEventRepository) Remove(ctx context.Context, eventIDs []int64) error {
	var success bool
	start := time.Now()
	ctx, span := r.observ.StartSpan(ctx, "bookEventRepository.remove")
	span.SetAttributes([]observability.Attribute{{Key: "eventIDs", Value: eventIDs}})

	defer span.End()

	defer func() {
		duration := time.Since(start).Seconds()
		r.observ.RecordDatabaseQuery(ctx, "delete", "book_event", duration, success)
	}()

	query, args, err := r.builder.Delete("book_event").
		Where(sq.And{sq.Eq{"id": eventIDs}, sq.Eq{"status": entities.EventStatusLock}}).
		ToSql()
	if err != nil {
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "toSql.failed", Value: true}})

		return errors.Wrap(err, "bookEventPostgres.Remove: building query")
	}

	res, err := r.querier.ExecContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "scan.failed", Value: true}})

		return errors.Wrap(err, "bookEventPostgres.Remove: executing query")
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "rowsAffected.failed", Value: true}})

		return errors.Wrap(err, "bookEventPostgres.Remove: getting rows affected")
	}

	if rowsAffected != int64(len(eventIDs)) {
		span.SetAttributes([]observability.Attribute{{Key: "len.bookEvent.noEqual.failed", Value: true}})

		return errors.New(fmt.Sprintf("bookEventPostgres.Remove: expected rowsAffected %d, actual %d", len(eventIDs), rowsAffected))
	}

	success = true
	return nil
}
