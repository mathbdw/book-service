package book

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/mathbdw/book/internal/domain/entities"
	"github.com/mathbdw/book/internal/errors"
	"github.com/mathbdw/book/internal/interfaces/repositories"
	"github.com/mathbdw/book/mocks"
	"github.com/stretchr/testify/assert"
)

func createMockUsecaseObservability(ctrl *gomock.Controller) *mocks.MockUsecaseObservability {
	observ := mocks.NewMockUsecaseObservability(ctrl)
	mockSpan := mocks.NewMockSpan(ctrl)

	observ.EXPECT().StartSpan(gomock.Any(), gomock.Any()).Return(context.Background(), mockSpan).AnyTimes()
	observ.EXPECT().RecordBookCreated(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockSpan.EXPECT().End().AnyTimes()
	mockSpan.EXPECT().RecordError(gomock.Any()).AnyTimes()
	mockSpan.EXPECT().SetAttributes(gomock.Any()).AnyTimes()

	return observ
}

func TestBook_Create_ErrorRepoBook(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uowMock := mocks.NewMockUnitOfWork(ctrl)
	bookMock := mocks.NewMockBookRepository(ctrl)
	bookEventMock := mocks.NewMockBookEventRepository(ctrl)
	observUsecase := createMockUsecaseObservability(ctrl)
	us := NewAddBookUsecase(uowMock, observUsecase)

	book := entities.Book{Title: "Test", Description: "Test Desc", Genre: "Test Genre", Year: 2019}

	ctx := context.Background()
	uowMock.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(repo *repositories.Repository) error) error {
			bookMock.EXPECT().
				Create(ctx, book).
				Return(int64(0), errors.New("error repoBook"))

			bookEventMock.EXPECT().
				Create(ctx, gomock.Any()).
				Return(int64(1), nil).
				Times(0)

			repo := &repositories.Repository{
				Book:      bookMock,
				BookEvent: bookEventMock,
			}

			return fn(repo)
		})

	err := us.Execute(ctx, book)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "addBookUsecases.Execute: failed to save book")
}

func TestBook_Create_ErrorRepoBookEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uowMock := mocks.NewMockUnitOfWork(ctrl)
	bookMock := mocks.NewMockBookRepository(ctrl)
	bookEventMock := mocks.NewMockBookEventRepository(ctrl)
	observUsecase := createMockUsecaseObservability(ctrl)
	us := NewAddBookUsecase(uowMock, observUsecase)

	book := entities.Book{Title: "Test", Description: "Test Desc", Genre: "Test Genre", Year: 2019}

	ctx := context.Background()
	uowMock.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(repo *repositories.Repository) error) error {
			bookMock.EXPECT().
				Create(ctx, book).
				Return(int64(1), nil)

			bookEventMock.EXPECT().
				Create(ctx, gomock.Any()).
				Return(int64(0), errors.New("error repoBook"))

			repo := &repositories.Repository{
				Book:      bookMock,
				BookEvent: bookEventMock,
			}

			return fn(repo)
		})

	err := us.Execute(ctx, book)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "addBookUsecases.Execute: create book event")
}

func TestBook_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uowMock := mocks.NewMockUnitOfWork(ctrl)
	bookMock := mocks.NewMockBookRepository(ctrl)
	bookEventMock := mocks.NewMockBookEventRepository(ctrl)
	observUsecase := createMockUsecaseObservability(ctrl)
	us := NewAddBookUsecase(uowMock, observUsecase)

	book := entities.Book{Title: "Test", Description: "Test Desc", Genre: "Test Genre", Year: 2019}

	ctx := context.Background()
	uowMock.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(repo *repositories.Repository) error) error {
			bookMock.EXPECT().
				Create(ctx, book).
				Return(int64(1), nil)

			bookEventMock.EXPECT().
				Create(ctx, gomock.Any()).
				Return(int64(1), nil)

			repo := &repositories.Repository{
				Book:      bookMock,
				BookEvent: bookEventMock,
			}

			return fn(repo)
		})

	err := us.Execute(ctx, book)

	assert.NoError(t, err)
}
