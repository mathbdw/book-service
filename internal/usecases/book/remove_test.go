package book

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/mathbdw/book/internal/domain/entities"
	errs "github.com/mathbdw/book/internal/errors"
	"github.com/mathbdw/book/internal/interfaces/repositories"
	"github.com/mathbdw/book/mocks"
)

func TestBook_Remove_ErrorBook(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uowMock := mocks.NewMockUnitOfWork(ctrl)
	bookMock := mocks.NewMockBookRepository(ctrl)
	bookEventMock := mocks.NewMockBookEventRepository(ctrl)
	//observUsecase - add_book_test.go
	observUsecase := createMockUsecaseObservability(ctrl)
	ctx := context.Background()

	uowMock.EXPECT().Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(repo *repositories.Repository) error) error {
			bookMock.EXPECT().
				Remove(ctx, gomock.Any()).
				Return(errs.ErrNotFound)

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

	us := NewRemoveBookUsecase(uowMock, observUsecase)
	err := us.Execute(ctx, []int64{1, 2})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "RemoveBookUsecase.Execute: remove Book")
}

func TestBook_Remove_ErrorBookEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uowMock := mocks.NewMockUnitOfWork(ctrl)
	bookMock := mocks.NewMockBookRepository(ctrl)
	bookEventMock := mocks.NewMockBookEventRepository(ctrl)
	//observUsecase - add_book_test.go
	observUsecase := createMockUsecaseObservability(ctrl)
	ctx := context.Background()

	uowMock.EXPECT().Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(repo *repositories.Repository) error) error {
			bookMock.EXPECT().
				Remove(ctx, gomock.Any()).
				Return(nil)

			bookEventMock.EXPECT().
				Create(ctx, gomock.Any()).
				Return(int64(0), errs.New("error"))

			repo := &repositories.Repository{
				Book:      bookMock,
				BookEvent: bookEventMock,
			}

			return fn(repo)
		})

	us := NewRemoveBookUsecase(uowMock, observUsecase)
	err := us.Execute(ctx, []int64{1, 2})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "RemoveBookUsecase.Execute: create Book Event")
}

func TestBook_Remove_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uowMock := mocks.NewMockUnitOfWork(ctrl)
	bookMock := mocks.NewMockBookRepository(ctrl)
	bookEventMock := mocks.NewMockBookEventRepository(ctrl)
	//observUsecase - add_book_test.go
	observUsecase := createMockUsecaseObservability(ctrl)
	ctx := context.Background()

	uowMock.EXPECT().Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(repo *repositories.Repository) error) error {
			ids := []int64{1, 2}
			bookMock.EXPECT().
				Remove(ctx, ids).
				Return(nil)

			for _, id := range ids {
				bookEvent := entities.BookEvent{BookId: int64(id), Type: entities.Deleted, Status: entities.EventStatusNew}
				bookEventMock.EXPECT().
					Create(ctx, bookEvent).
					Return(int64(1), nil)
			}

			repo := &repositories.Repository{
				Book:      bookMock,
				BookEvent: bookEventMock,
			}

			return fn(repo)
		})

	us := NewRemoveBookUsecase(uowMock, observUsecase)
	err := us.Execute(ctx, []int64{1, 2})

	assert.NoError(t, err)
}
