package handlers

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/mathbdw/book/internal/usecases/book"
	"github.com/mathbdw/book/internal/interfaces/repositories"

	"github.com/mathbdw/book/internal/domain/entities"
	errs "github.com/mathbdw/book/internal/errors"
	"github.com/mathbdw/book/mocks"
	pb "github.com/mathbdw/book/proto"
	"github.com/stretchr/testify/assert"
)

func removeMockUC(ctrl *gomock.Controller, uowRepo repositories.UnitOfWork, bookRepo repositories.BookRepository) *book.BookUsecases{
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

func TestBook_Delete_ErrorValidate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uowRepo := mocks.NewMockUnitOfWork(ctrl)
	bookRepo := mocks.NewMockBookRepository(ctrl)
	observHandler := createMockHandlerObservability(ctrl)
	uc := removeMockUC(ctrl, uowRepo, bookRepo)
	bookHandler := &BookHandler{uc: uc, observ: observHandler}
	ctx := context.Background()

	res, err := bookHandler.Delete(ctx, &pb.BookGetRequest{BookId: []int64{}})

	assert.Nil(t, res)
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestBook_Delete_ErrorUsecase_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uowRepo := mocks.NewMockUnitOfWork(ctrl)
	bookRepo := mocks.NewMockBookRepository(ctrl)
	bookEventRepo := mocks.NewMockBookEventRepository(ctrl)
	observHandler := createMockHandlerObservability(ctrl)
	uc := removeMockUC(ctrl, uowRepo, bookRepo)
	bookHandler := &BookHandler{uc: uc, observ: observHandler}
	ctx := context.Background()

	uowRepo.EXPECT().Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(repo *repositories.Repository) error) error {
			bookRepo.EXPECT().
				Remove(ctx, gomock.Any()).
				Return(errs.ErrNotFound)

			bookEventRepo.EXPECT().
				Create(ctx, gomock.Any()).
				Return(int64(1), nil).
				Times(0)

			repo := &repositories.Repository{
				Book:      bookRepo,
				BookEvent: bookEventRepo,
			}

			return fn(repo)
		})

	res, err := bookHandler.Delete(ctx, &pb.BookGetRequest{BookId: []int64{1}})

	assert.Nil(t, res)
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Contains(t, err.Error(), "not found")
}

func TestBook_Delete_ErrorUsecase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uowRepo := mocks.NewMockUnitOfWork(ctrl)
	bookRepo := mocks.NewMockBookRepository(ctrl)
	bookEventRepo := mocks.NewMockBookEventRepository(ctrl)
	observHandler := createMockHandlerObservability(ctrl)
	uc := removeMockUC(ctrl, uowRepo, bookRepo)
	bookHandler := &BookHandler{uc: uc, observ: observHandler}
	ctx := context.Background()

	uowRepo.EXPECT().Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(repo *repositories.Repository) error) error {
			bookRepo.EXPECT().
				Remove(ctx, gomock.Any()).
				Return(errs.New("error"))

			bookEventRepo.EXPECT().
				Create(ctx, gomock.Any()).
				Return(int64(1), nil).
				Times(0)

			repo := &repositories.Repository{
				Book:      bookRepo,
				BookEvent: bookEventRepo,
			}

			return fn(repo)
		})

	res, err := bookHandler.Delete(ctx, &pb.BookGetRequest{BookId: []int64{1}})

	assert.Nil(t, res)
	assert.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestBook_Delete_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uowRepo := mocks.NewMockUnitOfWork(ctrl)
	bookRepo := mocks.NewMockBookRepository(ctrl)
	bookEventRepo := mocks.NewMockBookEventRepository(ctrl)
	observHandler := createMockHandlerObservability(ctrl)
	uc := removeMockUC(ctrl, uowRepo, bookRepo)
	bookHandler := &BookHandler{uc: uc, observ: observHandler}
	ctx := context.Background()

	uowRepo.EXPECT().Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(repo *repositories.Repository) error) error {
			ids := []int64{1, 2}
			bookRepo.EXPECT().
				Remove(ctx, ids).
				Return(nil)

			for _, id := range ids {
				bookEvent := entities.BookEvent{BookId: int64(id), Type: entities.Deleted, Status: entities.EventStatusNew}
				bookEventRepo.EXPECT().
					Create(ctx, bookEvent).
					Return(int64(1), nil)
			}

			repo := &repositories.Repository{
				Book:      bookRepo,
				BookEvent: bookEventRepo,
			}

			return fn(repo)
		})

	res, err := bookHandler.Delete(ctx, &pb.BookGetRequest{BookId: []int64{1, 2}})

	assert.Nil(t, err)
	assert.Equal(t, &emptypb.Empty{}, res)
}
