package handlers

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mathbdw/book/internal/domain/entities"
	errs "github.com/mathbdw/book/internal/errors"
	"github.com/mathbdw/book/internal/interfaces/repositories"
	"github.com/mathbdw/book/internal/usecases/book"
	"github.com/mathbdw/book/mocks"
	pb "github.com/mathbdw/book/proto"
	"github.com/stretchr/testify/assert"
)

func getMockUC(ctrl *gomock.Controller, bookRepo repositories.BookRepository) *book.BookUsecases{
	uowRepo := mocks.NewMockUnitOfWork(ctrl)
	observUsecase := createMockUsecaseObservability(ctrl)

	addUC := book.NewAddBookUsecase(uowRepo, observUsecase)
	getUC := book.NewGetBookUsecase(bookRepo, observUsecase)
	listUC := book.NewListBookUsecase(bookRepo, observUsecase)
	removeUC := book.NewRemoveBookUsecase(uowRepo, observUsecase)

	return book.New(
		book.WithAddBookUsecase(addUC),
		book.WithGetBookUsecase(getUC),
		book.WithListBookUsecase(listUC),
		book.WithRemoveBookUsecase(removeUC),
	)
}

func TestBook_GetByIDs_ErrorValidate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bookRepo := mocks.NewMockBookRepository(ctrl)
	//createMockObservability - add_book_test.go
	observHandler := createMockHandlerObservability(ctrl)
	uc := getMockUC(ctrl, bookRepo)
	bookHandler := &BookHandler{uc: uc, observ: observHandler}
	ctx := context.Background()

	tests := []struct {
		name    string
		dataReq []int64
		errStr  string
	}{
		{"UniqueItem", []int64{1, 1}, "repeated value must contain unique items"},
		{"EmptyItem", []int64{}, "value must contain between"},
		{"Gte1", []int64{-1}, "value must be greater than or equal to"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := bookHandler.GetByIDs(ctx, &pb.BookGetRequest{
				BookId: tt.dataReq,
			})

			assert.Nil(t, res)
			assert.Error(t, err)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
			assert.Contains(t, err.Error(), tt.errStr)
		})
	}
}

func TestBook_GetByIDs_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bookRepo := mocks.NewMockBookRepository(ctrl)
	//createMockObservability - add_book_test.go
	observHandler := createMockHandlerObservability(ctrl)
	uc := getMockUC(ctrl, bookRepo)
	bookHandler := &BookHandler{uc: uc, observ: observHandler}
	ctx := context.Background()

	bookRepo.EXPECT().
		GetByIDs(gomock.Any(), []int64{1, 2}).
		Return([]entities.Book{}, errs.New("not found"))

	res, err := bookHandler.GetByIDs(ctx, &pb.BookGetRequest{
		BookId: []int64{1, 2},
	})

	assert.Nil(t, res)
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestBook_GetByIDs_ErrorUsecase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bookRepo := mocks.NewMockBookRepository(ctrl)
	//createMockObservability - add_book_test.go
	observHandler := createMockHandlerObservability(ctrl)
	uc := getMockUC(ctrl, bookRepo)
	bookHandler := &BookHandler{uc: uc, observ: observHandler}
	ctx := context.Background()

	bookRepo.EXPECT().
		GetByIDs(gomock.Any(), []int64{1, 2}).
		Return([]entities.Book{}, errs.New("error"))

	res, err := bookHandler.GetByIDs(ctx, &pb.BookGetRequest{
		BookId: []int64{1, 2},
	})

	assert.Nil(t, res)
	assert.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestBook_GetByIDs_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bookRepo := mocks.NewMockBookRepository(ctrl)
	//createMockObservability - add_book_test.go
	observHandler := createMockHandlerObservability(ctrl)
	uc := getMockUC(ctrl, bookRepo)
	bookHandler := &BookHandler{uc: uc, observ: observHandler}
	ctx := context.Background()

	bookRepo.EXPECT().
		GetByIDs(gomock.Any(), []int64{1, 2}).
		Return([]entities.Book{
			{ID: 1, Title: "Title1", Description: "Desc1", Year: 1900, Genre: "Genre1"},
			{ID: 2, Title: "Title2", Description: "Desc2", Year: 1902, Genre: "Genre2"},
		}, nil)

	res, err := bookHandler.GetByIDs(ctx, &pb.BookGetRequest{
		BookId: []int64{1, 2},
	})

	assert.Equal(t, 2, len(res.Book))
	assert.NoError(t, err)
}
