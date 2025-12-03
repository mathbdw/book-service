package postgres

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/mathbdw/book/internal/interfaces/observability"
	"github.com/mathbdw/book/internal/interfaces/repositories"
	"github.com/pkg/errors"
)

type unitOfWork struct {
	db      *sqlx.DB
	builder sq.StatementBuilderType

	observ observability.RepositoryObservability
}

// NewUnitOfWork - Constructor Unit of Work
func NewUnitOfWork(db *sqlx.DB, builder sq.StatementBuilderType, observ observability.RepositoryObservability) repositories.UnitOfWork {
	return &unitOfWork{db: db, builder: builder, observ: observ}
}

func (uow *unitOfWork) Do(ctx context.Context, fn func(repo *repositories.Repository) error) error {
	tx, err := uow.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "uowPostgres.Do: failed to begin transaction")
	}
	defer tx.Rollback()

	repos := &repositories.Repository{
		Book:      NewBookRepository(tx, uow.builder, uow.observ),
		BookEvent: NewBookEventRepository(tx, uow.builder, uow.observ),
	}

	err = fn(repos)
	if err != nil {
		return errors.Wrap(err, "uowPostgres.Do: аn error occurred while executing the function")
	}

	return tx.Commit()
}
