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

func TestBook_List_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bookMock := mocks.NewMockBookRepository(ctrl)
	//observUsecase - add_book_test.go
	observUsecase := createMockUsecaseObservability(ctrl)
	ctx := context.Background()

	bookMock.EXPECT().
		List(gomock.Any(), gomock.Any()).
		Return(nil, errs.New("not found"))

	us := NewListBookUsecase(bookMock, observUsecase)
	responseBooks, err := us.Execute(ctx, entities.PaginationParams{})

	assert.Empty(t, responseBooks)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "listBookUsecase.Execute: list books")
	assert.Equal(t, errors.Is(err, errs.ErrNotFound), true)
}

func TestBook_List_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bookMock := mocks.NewMockBookRepository(ctrl)
	//observUsecase - add_book_test.go
	observUsecase := createMockUsecaseObservability(ctrl)
	ctx := context.Background()

	bookMock.EXPECT().
		List(gomock.Any(), gomock.Any()).
		Return(&entities.ResponseBooks{Data: []entities.Book{{ID: 1, Title: "Test"}}, PageInfo: entities.PageInfo{}}, nil)

	us := NewListBookUsecase(bookMock, observUsecase)
	responseBooks, err := us.Execute(ctx, entities.PaginationParams{})

	assert.Equal(t, len(responseBooks.Data), 1)
	assert.NoError(t, err)
}
