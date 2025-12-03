package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/mathbdw/book/internal/domain/entities"
	errs "github.com/mathbdw/book/internal/errors"
)

func TestBook_Create_ErrorScan(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	//createMockMockRepositoryObservability - book_event_postgres_test.go
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO book (description,genre,title,year) VALUES ($1,$2,$3,$4) RETURNING id")).
		WithArgs(
			"Test Description",
			"Test genre",
			"Test Book",
			2021,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("not_an_integer"))

	bookId, err := repo.Create(ctx, entities.Book{
		Title:       "Test Book",
		Description: "Test Description",
		Year:        2021,
		Genre:       "Test genre",
	})

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bookPostgres.Create: error scanning")
	assert.Equal(t, int64(0), bookId)
}

func TestBook_Create_Success(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	//createMockMockRepositoryObservability - book_event_postgres_test.go
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO book (description,genre,title,year) VALUES ($1,$2,$3,$4) RETURNING id")).
		WithArgs(
			"Test Description",
			"Test genre",
			"Test Book",
			2021,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(100))

	bookId, err := repo.Create(ctx, entities.Book{
		Title:       "Test Book",
		Description: "Test Description",
		Year:        2021,
		Genre:       "Test genre",
	})

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Nil(t, err)
	assert.Equal(t, bookId, int64(100))
}

func TestBook_GetById_ErrorQuery(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	//createMockMockRepositoryObservability - book_event_postgres_test.go
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM book WHERE (id IN ($1,$2) AND removed = $3)")).
		WithArgs(1, 2, false).
		WillReturnError(sql.ErrNoRows)

	books, err := repo.GetByIDs(ctx, []int64{1, 2})

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.True(t, errors.Is(err, sql.ErrNoRows), fmt.Sprintf("Expected sql.ErrNoRows, got: %v", err))
	assert.Contains(t, err.Error(), "bookPostgres.GetByIds: error query")
	assert.Empty(t, books)
}

func TestBook_GetById_ErrorScan(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	//createMockMockRepositoryObservability - book_event_postgres_test.go
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM book WHERE (id IN ($1,$2) AND removed = $3)")).
		WithArgs(1, 2, false).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "title", "genre", "description", "year"}).
				AddRow("", "Test Title", "Test Description", "Test Book", 2021),
		)

	books, err := repo.GetByIDs(ctx, []int64{1, 2})

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bookPostgres.GetByIds: error scan")
	assert.Empty(t, books)
}

func TestBook_GetById_NotFound(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	//createMockMockRepositoryObservability - book_event_postgres_test.go
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM book WHERE (id IN ($1,$2) AND removed = $3)")).
		WithArgs(1, 2, false).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "title", "genre", "description", "year"}),
		)

	books, err := repo.GetByIDs(ctx, []int64{1, 2})

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), errs.ErrNotFound.Error())
	assert.Empty(t, books)
}

func TestBook_GetById_Success(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	//createMockMockRepositoryObservability - book_event_postgres_test.go
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM book WHERE (id IN ($1,$2) AND removed = $3)")).
		WithArgs(1, 2, false).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "title", "description", "year", "genre"}).
				AddRow(1, "Test Book", "Test Description", 2021, "Test genre").
				AddRow(2, "Test Book2", "Test Description2", 2022, "Test genre2"),
		)

	models, err := repo.GetByIDs(ctx, []int64{1, 2})

	assert.Nil(t, err)
	assert.Equal(t, len(models), 2, "Expected models to be equal")
	assert.Equal(t, "Test Book2", (models)[1].Title)
	assert.Equal(t, "Test Description", (models)[0].Description)
	assert.Equal(t, 2022, (models)[1].Year)
	assert.Equal(t, "Test genre", (models)[0].Genre)
}

func TestBook_List_ErrorQuery(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	//createMockMockRepositoryObservability - book_event_postgres_test.go
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM book ORDER BY id asc LIMIT 2")).
		WithoutArgs().
		WillReturnError(sql.ErrNoRows)

	respBooks, err := repo.List(ctx, entities.PaginationParams{
		Limit:     1,
		SortOrder: entities.SortOrderTypeAsc,
		SortBy:    entities.CursorTypeBookID,
	})

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.True(t, errors.Is(err, sql.ErrNoRows), "Expected error to wrap sql.ErrNoRows, got: %v", err)
	assert.Contains(t, err.Error(), "bookPostgres.List: error query")
	assert.Nil(t, respBooks)
}

func TestBook_List_ErrorScan(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	//createMockMockRepositoryObservability - book_event_postgres_test.go
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM book ORDER BY id asc LIMIT 2")).
		WithoutArgs().
		WillReturnRows(mock.NewRows([]string{"id", "title", "genre", "description", "year"}).
			AddRow("", "Test Book", "Test Description", 2021, "Test genre"),
		)

	respBooks, err := repo.List(ctx, entities.PaginationParams{
		Limit:     1,
		SortOrder: entities.SortOrderTypeAsc,
		SortBy:    entities.CursorTypeBookID,
	})

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bookPostgres.List: error scan")
	assert.Nil(t, respBooks)
}

func TestBook_List_EmptyScan(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	//createMockMockRepositoryObservability - book_event_postgres_test.go
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM book ORDER BY id asc LIMIT 2")).
		WithoutArgs().
		WillReturnRows(mock.NewRows([]string{"id", "title", "genre", "description", "year"}))

	respBooks, err := repo.List(ctx, entities.PaginationParams{
		Limit:     1,
		SortOrder: entities.SortOrderTypeAsc,
		SortBy:    entities.CursorTypeBookID,
	})

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bookPostgres.List: books not found")
	assert.Nil(t, respBooks)
}

func TestBook_List_ErrorCreatePageInfo(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	//createMockMockRepositoryObservability - book_event_postgres_test.go
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM book WHERE title > $1 ORDER BY title asc LIMIT 2")).
		WithArgs("title").
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "title"}).
				AddRow(1, "").
				AddRow(1, "Test title2"),
		)

	respBooks, err := repo.List(ctx, entities.PaginationParams{
		Cursor:    &entities.Cursor{Value: "title"},
		Limit:     1,
		SortBy:    entities.CursorTypeBookTitle,
		SortOrder: entities.SortOrderTypeAsc,
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bookPostgres.List: error createPageInfo")
	assert.Nil(t, respBooks)
}

var multipleRows = [][]driver.Value{
	{1, "Book 1", "Desc 1", "Genre 1", 2021, time.Date(2021, time.January, 1, 8, 0, 0, 0, time.UTC)},
	{2, "Book 2", "Desc 2", "Genre 2", 2022, time.Date(2021, time.January, 1, 9, 0, 0, 0, time.UTC)},
	{3, "Book 3", "Desc 3", "Genre 3", 2022, time.Date(2021, time.January, 1, 10, 0, 0, 0, time.UTC)},
	{4, "Book 4", "Desc 4", "Genre 4", 2023, time.Date(2021, time.January, 1, 11, 0, 0, 0, time.UTC)},
	{5, "Book 5", "Desc 5", "Genre 5", 2024, time.Date(2021, time.January, 1, 12, 0, 0, 0, time.UTC)},
}

func TestBook_List(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	//createMockMockRepositoryObservability - book_event_postgres_test.go
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM book WHERE id > $1 ORDER BY id asc LIMIT 3")).
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "title", "description", "genre", "year", "created_at"}).
				AddRow(multipleRows[1]...).
				AddRow(multipleRows[2]...).
				AddRow(multipleRows[3]...),
		)

	responseBook, err := repo.List(ctx, entities.PaginationParams{
		Cursor:    &entities.Cursor{Value: 1},
		Limit:     2,
		SortBy:    entities.CursorTypeBookID,
		SortOrder: entities.SortOrderTypeAsc,
	})

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.NoError(t, err, "Expected no error")
	assert.Equal(t, 2, len((*responseBook).Data))
	assert.NotNil(t, responseBook.PageInfo)
	assert.Equal(t, int64(2), responseBook.Data[0].ID)
}

func TestBook_ListLess(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	//createMockMockRepositoryObservability - book_event_postgres_test.go
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM book WHERE id > $1 ORDER BY id asc LIMIT 5")).
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "title", "description", "genre", "year", "created_at"}).
				AddRow(multipleRows[1]...).
				AddRow(multipleRows[2]...).
				AddRow(multipleRows[3]...),
		)

	responseBook, err := repo.List(ctx, entities.PaginationParams{
		Cursor:    &entities.Cursor{Value: 1},
		Limit:     4,
		SortBy:    entities.CursorTypeBookID,
		SortOrder: entities.SortOrderTypeAsc,
	})

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.NoError(t, err, "Expected no error")
	assert.Equal(t, 3, len((*responseBook).Data))
	assert.NotNil(t, responseBook.PageInfo)
	assert.Equal(t, int64(2), responseBook.Data[0].ID)
}

func TestBook_Remove_ErrorNoRows(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	//createMockMockRepositoryObservability - book_event_postgres_test.go
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE book SET removed = $1, updated_at = $2 WHERE id IN ($3,$4)")).
		WithArgs(true, sqlmock.AnyArg(), 1, 2).
		WillReturnError(sql.ErrNoRows)

	err = repo.Remove(ctx, []int64{1, 2})

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Error(t, err)
	assert.True(t, errors.Is(err, sql.ErrNoRows), fmt.Sprintf("Expected sql.ErrNoRows, got: %v", err))
}

// Create custom ErrorResult
type ErrorResult struct{}

func (r *ErrorResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (r *ErrorResult) RowsAffected() (int64, error) {
	return 0, errors.New("rows affected error")
}

func TestBook_Remove_ErrorGetAffectedRows(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	//createMockMockRepositoryObservability - book_event_postgres_test.go
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE book SET removed = $1, updated_at = $2 WHERE id IN ($3,$4)")).
		WithArgs(true, sqlmock.AnyArg(), 1, 2).
		WillReturnResult(&ErrorResult{})

	err = repo.Remove(ctx, []int64{1, 2})

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bookPostgres.Remove: error get affected rows")
}

func TestBook_Remove_ErrorNotEquilRowsAffected(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	//createMockMockRepositoryObservability - book_event_postgres_test.go
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE book SET removed = $1, updated_at = $2 WHERE id IN ($3,$4)")).
		WithArgs(true, sqlmock.AnyArg(), 1, 2).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Remove(ctx, []int64{1, 2})

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bookPostgres.Remove: expected rowsAffected")
}

func TestBook_Remove_Success(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	//createMockMockRepositoryObservability - book_event_postgres_test.go
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE book SET removed = $1, updated_at = $2 WHERE id IN ($3,$4)")).
		WithArgs(true, sqlmock.AnyArg(), 1, 2).
		WillReturnResult(sqlmock.NewResult(0, 2))

	err = repo.Remove(ctx, []int64{1, 2})

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.NoError(t, err)
}
