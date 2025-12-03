package book

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/mathbdw/book/internal/domain/entities"
	errs "github.com/mathbdw/book/internal/errors"
	"github.com/mathbdw/book/mocks"
	"github.com/stretchr/testify/assert"
)

func TestBook_GetByIDs_ErrorNoRows(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bookMock := mocks.NewMockBookRepository(ctrl)
	//observUsecase - add_book_test.go
	observUsecase := createMockUsecaseObservability(ctrl)
	ctx := context.Background()

	bookMock.EXPECT().
		GetByIDs(gomock.Any(), []int64{1, 2}).
		Return([]entities.Book{}, errs.New("not found"))

	us := NewGetBookUsecase(bookMock, observUsecase)
	books, err := us.GetByIDs(ctx, []int64{1, 2})

	assert.Empty(t, books)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GetBookUsecase.GetByIds: get books")
	assert.Equal(t, errors.Is(err, errs.ErrNotFound), true)
}

func TestBook_GetByIDs_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bookMock := mocks.NewMockBookRepository(ctrl)
	//observUsecase - add_book_test.go
	observUsecase := createMockUsecaseObservability(ctrl)
	ctx := context.Background()

	bookMock.EXPECT().
		GetByIDs(gomock.Any(), []int64{1, 2}).
		Return([]entities.Book{
			{ID: 1, Title: "Title", Description: "Desc", Year: 1900, Genre: "Genre"},
			{ID: 2, Title: "Title 2", Description: "Desc 2", Year: 2000, Genre: "Genre"},
		}, nil)

	us := NewGetBookUsecase(bookMock, observUsecase)
	books, err := us.GetByIDs(ctx, []int64{1, 2})

	assert.NoError(t, err)
	assert.Equal(t, len(books), 2)
}
