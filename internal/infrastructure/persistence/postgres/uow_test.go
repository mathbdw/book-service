package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/mathbdw/book/internal/domain/entities"
	"github.com/mathbdw/book/internal/interfaces/repositories"
)

func TestUnitOfWork_Do_ErrorBeginTxx(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	//createMockMockRepositoryObservability - book_event_postgres_test.go
	observ := createMockMockRepositoryObservability(ctrl)
	uow := NewUnitOfWork(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectBegin().WillReturnError(errors.New("error"))
	err = uow.Do(ctx, func(repo *repositories.Repository) error {
		return nil
	})

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "uowPostgres.Do: failed to begin transaction")
}

func TestUnitOfWork_Do_ErrorExecutingFunction(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	//createMockMockRepositoryObservability - book_event_postgres_test.go
	observ := createMockMockRepositoryObservability(ctrl)
	uow := NewUnitOfWork(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectBegin()

	book := entities.Book{
		Title:       "Test Book",
		Description: "Test Description",
		Year:        1904,
		Genre:       "Test Genre",
	}
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO book (description,genre,title,year) VALUES ($1,$2,$3,$4) RETURNING id`)).
		WithArgs("Test Description", "Test Genre", "Test Book", 1904).
		WillReturnError(errors.New("error"))

	mock.ExpectRollback()

	err = uow.Do(ctx, func(repo *repositories.Repository) error {
		book.ID, err = repo.Book.Create(ctx, book)
		if err != nil {
			return err
		}

		return nil
	})

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "uowPostgres.Do: аn error occurred while executing the function")
}

func TestUnitOfWork_Do_Success(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	//createMockMockRepositoryObservability - book_event_postgres_test.go
	observ := createMockMockRepositoryObservability(ctrl)
	uow := NewUnitOfWork(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectBegin()

	book := entities.Book{
		Title:       "Test Book",
		Description: "Test Description",
		Year:        1904,
		Genre:       "Test Genre",
	}
	strBook, _ := json.Marshal(book)

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO book (description,genre,title,year) VALUES ($1,$2,$3,$4) RETURNING id`)).
		WithArgs("Test Description", "Test Genre", "Test Book", 1904).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO book_event (book_id,payload,status,type) VALUES ($1,$2,$3,$4) RETURNING id`)).
		WithArgs(1, strBook, entities.EventStatusNew, entities.Created).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectCommit()

	err = uow.Do(ctx, func(repo *repositories.Repository) error {
		book.ID, err = repo.Book.Create(ctx, book)
		if err != nil {
			return err
		}

		event := entities.BookEvent{BookId: book.ID, Type: entities.Created, Status: entities.EventStatusNew, Payload: strBook}
		event.ID, err = repo.BookEvent.Create(ctx, event)
		if err != nil {
			return err
		}

		return nil
	})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
