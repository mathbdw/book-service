package postgres

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/mathbdw/book/internal/domain/entities"
	errs "github.com/mathbdw/book/internal/errors"
	"github.com/mathbdw/book/internal/interfaces/observability"
	"github.com/mathbdw/book/internal/interfaces/repositories"
)

type bookRepository struct {
	querier sqlx.ExtContext
	builder sq.StatementBuilderType
	service ServicePagination

	observ observability.RepositoryObservability
}

// NewBookRepository - Constructor BookRepository
func NewBookRepository(querier sqlx.ExtContext, builder sq.StatementBuilderType, observ observability.RepositoryObservability) repositories.BookRepository {
	return &bookRepository{
		querier: querier,
		builder: builder,
		service: NewService(),
		observ:  observ,
	}
}

// Create - Adds row
func (r *bookRepository) Create(ctx context.Context, book entities.Book) (int64, error) {
	var success bool
	start := time.Now()
	ctx, span := r.observ.StartSpan(ctx, "bookRepository.create")

	defer span.End()

	defer func() {
		duration := time.Since(start).Seconds()
		r.observ.RecordDatabaseQuery(ctx, "insert", "book", duration, success)
	}()

	data := map[string]interface{}{
		"title":       book.Title,
		"description": book.Description,
		"year":        book.Year,
		"genre":       book.Genre,
	}

	query, args, err := r.builder.Insert("book").SetMap(data).Suffix("RETURNING id").ToSql()
	if err != nil {
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "toSql.failed", Value: true}})

		return 0, errs.Wrap(err, "bookPostgres.Create: error builder")
	}

	var insertedID int64
	err = r.querier.QueryRowxContext(ctx, query, args...).Scan(&insertedID)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "scan.failed", Value: true}})

		return 0, errs.Wrap(err, "bookPostgres.Create: error scanning")
	}

	success = true
	return insertedID, nil
}

// GetByIDs - Returns books by IDs
func (r *bookRepository) GetByIDs(ctx context.Context, IDs []int64) ([]entities.Book, error) {
	var success bool
	start := time.Now()
	ctx, span := r.observ.StartSpan(ctx, "bookRepository.getByIDs")

	defer span.End()

	defer func() {
		duration := time.Since(start).Seconds()
		r.observ.RecordDatabaseQuery(ctx, "select", "book", duration, success)
	}()

	query, args, err := r.builder.Select("*").
		From("book").
		Where(sq.And{sq.Eq{"id": IDs}, sq.Eq{"removed": false}}).
		ToSql()
	if err != nil {
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "toSql.failed", Value: true}})

		return []entities.Book{}, errs.Wrap(err, "bookPostgres.GetByIds: error builder")
	}

	rows, err := r.querier.QueryxContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "queryxContext.failed", Value: true}})

		return []entities.Book{}, errs.Wrap(err, "bookPostgres.GetByIds: error query")
	}
	books := make([]entities.Book, 0, len(IDs))
	for rows.Next() {
		var book entities.Book
		err = rows.StructScan(&book)
		if err != nil {
			span.RecordError(err)
			span.SetAttributes([]observability.Attribute{{Key: "scan.failed", Value: true}})

			return []entities.Book{}, errs.Wrap(err, "bookPostgres.GetByIds: error scan")
		}
		books = append(books, book)
	}

	if len(books) == 0 {
		span.SetAttributes([]observability.Attribute{{Key: "len.book.zero", Value: true}})

		return []entities.Book{}, errs.Wrap(errs.ErrNotFound, "bookPostgres.GetByIds: len books")
	}

	success = true
	return books, nil
}

// List - Returns a list of books using pagination
func (r *bookRepository) List(ctx context.Context, params entities.PaginationParams) (*entities.ResponseBooks, error) {
	var success bool
	start := time.Now()
	ctx, span := r.observ.StartSpan(ctx, "bookRepository.list")

	defer span.End()

	defer func() {
		duration := time.Since(start).Seconds()
		r.observ.RecordDatabaseQuery(ctx, "select", "book", duration, success)
	}()

	limit := params.Limit + 1

	query := r.builder.Select("*").From("book")
	query = conditionBuilder(query, params)
	query = orderByBuilder(query, params)
	query = query.Limit(limit)

	sql, args, err := query.ToSql()
	if err != nil {
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "toSql.failed", Value: true}})

		return nil, errs.Wrap(err, "bookPostgres.List: error builder")
	}

	rows, err := r.querier.QueryxContext(ctx, sql, args...)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "queryxContext.failed", Value: true}})

		return nil, errs.Wrap(err, "bookPostgres.List: error query")
	}

	books := make([]entities.Book, 0, limit)
	for rows.Next() {
		var book entities.Book
		err = rows.StructScan(&book)
		if err != nil {
			span.RecordError(err)
			span.SetAttributes([]observability.Attribute{{Key: "scan.failed", Value: true}})

			return nil, errs.Wrap(err, "bookPostgres.List: error scan")
		}
		books = append(books, book)
	}

	if len(books) == 0 {
		span.SetAttributes([]observability.Attribute{{Key: "len.book.zero", Value: true}})

		return nil, fmt.Errorf("bookPostgres.List: books %s", errs.ErrNotFound)
	}

	paginatable := make([]entities.Paginatable, len(books))
	for i := range books {
		paginatable[i] = books[i]
	}

	pageInfo, err := r.service.CreatePageInfo(paginatable, params)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "createPageInfo.failed", Value: true}})

		return nil, errs.Wrap(err, "bookPostgres.List: error createPageInfo")
	}

	if uint64(len(books)) > params.Limit {
		books = books[:params.Limit]
	}

	success = true
	return &entities.ResponseBooks{
		Data:     books,
		PageInfo: pageInfo,
	}, nil
}

// Remove - Sets the field removed to true
func (r *bookRepository) Remove(ctx context.Context, IDs []int64) error {
	var success bool
	start := time.Now()
	ctx, span := r.observ.StartSpan(ctx, "bookRepository.remove")

	defer span.End()

	defer func() {
		duration := time.Since(start).Seconds()
		r.observ.RecordDatabaseQuery(ctx, "delete", "book", duration, success)
	}()

	query, args, err := r.builder.Update("book").
		Where(sq.Eq{"id": IDs}).
		Set("removed", true).
		Set("updated_at", time.Now().UTC()).
		ToSql()

	if err != nil {
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "toSql.failed", Value: true}})

		return errs.Wrap(err, "bookPostgres.Remove: error builder")
	}

	res, err := r.querier.ExecContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "execContext.failed", Value: true}})

		return errs.Wrap(err, "bookPostgres.Remove: error query")
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "rowsAffected.failed", Value: true}})

		return errs.Wrap(err, "bookPostgres.Remove: error get affected rows")
	}

	if rowsAffected != int64(len(IDs)) {
		span.SetAttributes([]observability.Attribute{{Key: "len.book.noEqual.failed", Value: true}})

		return errs.Wrap(errs.ErrNotFound, fmt.Sprintf("bookPostgres.Remove: expected rowsAffected %d, actual %d", len(IDs), rowsAffected))
	}

	success = true
	return nil
}
