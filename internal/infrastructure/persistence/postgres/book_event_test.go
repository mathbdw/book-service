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
	"github.com/mathbdw/book/mocks"
)

func createMockMockRepositoryObservability(ctrl *gomock.Controller) *mocks.MockRepositoryObservability {
	observ := mocks.NewMockRepositoryObservability(ctrl)
	mockSpan := mocks.NewMockSpan(ctrl)

	observ.EXPECT().StartSpan(gomock.Any(), gomock.Any()).Return(context.Background(), mockSpan).AnyTimes()
	observ.EXPECT().RecordDatabaseQuery(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockSpan.EXPECT().End().AnyTimes()
	mockSpan.EXPECT().RecordError(gomock.Any()).AnyTimes()
	mockSpan.EXPECT().SetAttributes(gomock.Any()).AnyTimes()

	return observ
}

func TestBookEvent_Create_ErrorExecutingQuery(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookEventRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO book_event (book_id,payload,status,type) VALUES ($1,$2,$3,$4) RETURNING id")).
		WithArgs(1, []byte("{\"title\": \"desc\", \"year\": 1900}"), entities.EventStatusNew, entities.Created).
		WillReturnError(sql.ErrNoRows)

	_, err = repo.Create(ctx, entities.BookEvent{
		BookId: 1, Type: entities.Created, Status: entities.EventStatusNew, Payload: []byte("{\"title\": \"desc\", \"year\": 1900}"),
	})

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bookEventPostgres.Create: scanning query")
}

func TestBookEvent_Create_Success(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookEventRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO book_event (book_id,payload,status,type) VALUES ($1,$2,$3,$4) RETURNING id")).
		WithArgs(1, []byte("{\"title\": \"desc\", \"year\": 1900}"), entities.EventStatusNew, entities.Created).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	res, err := repo.Create(ctx, entities.BookEvent{
		BookId: 1, Type: entities.Created, Status: entities.EventStatusNew, Payload: []byte("{\"title\": \"desc\", \"year\": 1900}"),
	})

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.NoError(t, err)
	assert.Equal(t, int64(1), res)
}

func TestBookEvent_Lock_ErrorExecQury(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookEventRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE book_event SET status = $1 WHERE status != $2 order by id asc limit $3")).
		WithArgs(entities.EventStatusLock, 2, 2).
		WillReturnError(sql.ErrNoRows)

	_, err = repo.Lock(ctx, 2)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "bookEventPostgres.Lock: executing query")
}

func TestBookEvent_Lock_ErrorNoRows(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookEventRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`
		WITH locked_event AS (SELECT id FROM book_event WHERE status <> $1 ORDER BY id ASC LIMIT 2 FOR UPDATE SKIP LOCKED)
		UPDATE book_event 
		SET status = $2
		WHERE id IN (SELECT id FROM locked_event)
		RETURNING id, book_id, type, payload
	`)).
		WithArgs(entities.EventStatusLock, 2).
		WillReturnError(sql.ErrNoRows)

	_, err = repo.Lock(ctx, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, sql.ErrNoRows), "Error should be sql.ErrNoRows")
}

func TestBookEvent_Lock_ErrorScanQuery(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookEventRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	var testSlice = [][]driver.Value{
		{1, "", entities.Created, entities.EventStatusLock, "{id: 1, title: test}", time.Now()},
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		WITH locked_event AS (SELECT id FROM book_event WHERE status <> $1 ORDER BY id ASC LIMIT 2 FOR UPDATE SKIP LOCKED)
		UPDATE book_event 
		SET status = $2
		WHERE id IN (SELECT id FROM locked_event)
		RETURNING id, book_id, type, payload
	`)).
		WithArgs(entities.EventStatusLock, 2).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "book_id", "type", "status", "payload", "updated_at"}).
				AddRow(testSlice[0]...),
		)

	_, err = repo.Lock(ctx, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bookEventPostgres.Lock: scanning row")
}

func TestBookEvent_Lock_ErrorIteration(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookEventRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	var testSlice = [][]driver.Value{
		{1, 32, entities.Created, entities.EventStatusLock, "{id: 1, title: test}", time.Now()},
		{2, 82, entities.Updated, entities.EventStatusLock, "{id: 5, title: test}", time.Now()},
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		WITH locked_event AS (SELECT id FROM book_event WHERE status <> $1 ORDER BY id ASC LIMIT 2 FOR UPDATE SKIP LOCKED)
		UPDATE book_event 
		SET status = $2
		WHERE id IN (SELECT id FROM locked_event)
		RETURNING id, book_id, type, payload
	`)).
		WithArgs(entities.EventStatusLock, 2).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "book_id", "type", "status", "payload", "updated_at"}).
				AddRow(testSlice[0]...).
				AddRow(testSlice[1]...).
				RowError(1, fmt.Errorf("iteration error")),
		)

	_, err = repo.Lock(ctx, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bookEventPostgres.Lock: errors during iteration")
}

func TestBookEvent_Lock_ErrorExpectLen(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookEventRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`
		WITH locked_event AS (SELECT id FROM book_event WHERE status <> $1 ORDER BY id ASC LIMIT 2 FOR UPDATE SKIP LOCKED)
		UPDATE book_event 
		SET status = $2
		WHERE id IN (SELECT id FROM locked_event)
		RETURNING id, book_id, type, payload
	`)).
		WithArgs(entities.EventStatusLock, 2).
		WillReturnError(errs.ErrNotFound)

	_, err = repo.Lock(ctx, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestBookEvent_Lock_Success(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookEventRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	var testSlice = [][]driver.Value{
		{1, 32, entities.Created, entities.EventStatusLock, "{id: 1, title: test}", time.Now()},
		{2, 82, entities.Updated, entities.EventStatusLock, "{id: 5, title: test}", time.Now()},
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		WITH locked_event AS (SELECT id FROM book_event WHERE status <> $1 ORDER BY id ASC LIMIT 2 FOR UPDATE SKIP LOCKED)
		UPDATE book_event 
		SET status = $2
		WHERE id IN (SELECT id FROM locked_event)
		RETURNING id, book_id, type, payload
	`)).
		WithArgs(entities.EventStatusLock, 2).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "book_id", "type", "status", "payload", "updated_at"}).
				AddRow(testSlice[0]...).
				AddRow(testSlice[1]...),
		)

	models, err := repo.Lock(ctx, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.NoError(t, err)
	assert.Equal(t, len(testSlice), len(models))
}

func TestBookEvent_Unlock_ErrorExecuting(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookEventRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE book_event SET status = $1 WHERE (id IN ($2,$3) AND status = $4)")).
		WithArgs(entities.EventStatusUnlock, 1, 2, entities.EventStatusLock).
		WillReturnError(fmt.Errorf("row error"))

	err = repo.Unlock(ctx, []int64{1, 2})

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bookEventPostgres.Unlock: executing query")
}

type ErrorResultBookEvent struct{}

func (r *ErrorResultBookEvent) LastInsertId() (int64, error) {
	return 0, nil
}

func (r *ErrorResultBookEvent) RowsAffected() (int64, error) {
	return 0, errors.New("rows affected error")
}

func TestBookEvent_Unlock_ErrorGetRowsAffected(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookEventRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE book_event SET status = $1 WHERE (id IN ($2,$3) AND status = $4)")).
		WithArgs(entities.EventStatusUnlock, 1, 2, entities.EventStatusLock).
		WillReturnResult(&ErrorResultBookEvent{})

	err = repo.Unlock(ctx, []int64{1, 2})

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bookEventPostgres.Unlock: getting rows affected")
}

func TestBookEvent_Unlock_ErrorRowsAffected(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookEventRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE book_event SET status = $1 WHERE (id IN ($2,$3) AND status = $4)")).
		WithArgs(entities.EventStatusUnlock, 1, 2, entities.EventStatusLock).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Unlock(ctx, []int64{1, 2})

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bookEventPostgres.Unlock: expected rowsAffected")
}

func TestBookEvent_Unlock_Success(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookEventRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE book_event SET status = $1 WHERE (id IN ($2,$3) AND status = $4)")).
		WithArgs(entities.EventStatusUnlock, 1, 2, entities.EventStatusLock).
		WillReturnResult(sqlmock.NewResult(1, 2))

	err = repo.Unlock(ctx, []int64{1, 2})

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.NoError(t, err)
}

func TestBookEvent_Remove_ErrorExecuting(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookEventRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM book_event WHERE (id IN ($1,$2) AND status = $3)")).
		WithArgs(1, 2, entities.EventStatusLock).
		WillReturnError(fmt.Errorf("row error"))

	err = repo.Remove(ctx, []int64{1, 2})

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bookEventPostgres.Remove: executing query")
}

func TestBookEvent_Remove_ErrorGetRowsAffected(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookEventRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM book_event WHERE (id IN ($1,$2) AND status = $3)")).
		WithArgs(1, 2, entities.EventStatusLock).
		WillReturnResult(&ErrorResultBookEvent{})

	err = repo.Remove(ctx, []int64{1, 2})

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bookEventPostgres.Remove: getting rows affected")
}

func TestBookEvent_Remove_ErrorRowsAffected(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookEventRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM book_event WHERE (id IN ($1,$2) AND status = $3)")).
		WithArgs(1, 2, entities.EventStatusLock).
		WillReturnResult(sqlmock.NewResult(1, 0))

	err = repo.Remove(ctx, []int64{1, 2})

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bookEventPostgres.Remove: expected rowsAffected")
}

func TestBookEvent_Remove_Success(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Error create mock")
	defer mockDB.Close()

	ctrl := gomock.NewController(t)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	observ := createMockMockRepositoryObservability(ctrl)
	repo := NewBookEventRepository(sqlxDB, builder, observ)
	ctx := context.Background()

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM book_event WHERE (id IN ($1,$2) AND status = $3)")).
		WithArgs(1, 2, entities.EventStatusLock).
		WillReturnResult(sqlmock.NewResult(1, 2))

	err = repo.Remove(ctx, []int64{1, 2})

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.NoError(t, err)
}
