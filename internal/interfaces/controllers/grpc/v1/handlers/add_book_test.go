package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/mathbdw/book/internal/usecases/book"
	"github.com/mathbdw/book/internal/domain/entities"
	"github.com/mathbdw/book/internal/interfaces/repositories"
	"github.com/mathbdw/book/mocks"
	pb "github.com/mathbdw/book/proto"
)

func createMockHandlerObservability(ctrl *gomock.Controller) *mocks.MockHandlerObservability {
	observ := mocks.NewMockHandlerObservability(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)
	mockSpan := mocks.NewMockSpan(ctrl)

	observ.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
	observ.EXPECT().StartSpan(gomock.Any(), gomock.Any()).Return(context.Background(), mockSpan).AnyTimes()
	observ.EXPECT().RecordHanderRequest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

	mockSpan.EXPECT().End().AnyTimes()
	mockSpan.EXPECT().RecordError(gomock.Any()).AnyTimes()
	mockSpan.EXPECT().SetAttributes(gomock.Any()).AnyTimes()

	return observ
}

func createMockUsecaseObservability(ctrl *gomock.Controller) *mocks.MockUsecaseObservability {
	observ := mocks.NewMockUsecaseObservability(ctrl)
	mockSpan := mocks.NewMockSpan(ctrl)

	// Настройка мока observability
	observ.EXPECT().StartSpan(gomock.Any(), gomock.Any()).Return(context.Background(), mockSpan).AnyTimes()
	observ.EXPECT().RecordBookCreated(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	// Настройка span методов
	mockSpan.EXPECT().End().AnyTimes()
	mockSpan.EXPECT().RecordError(gomock.Any()).AnyTimes()
	mockSpan.EXPECT().SetAttributes(gomock.Any()).AnyTimes()

	return observ
}

func createMockUC(ctrl *gomock.Controller, uowRepo repositories.UnitOfWork) *book.BookUsecases{
	bookRepo := mocks.NewMockBookRepository(ctrl)
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

func TestBook_Add_ErrorValidate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uowRepo := mocks.NewMockUnitOfWork(ctrl)
	observHandler := createMockHandlerObservability(ctrl)
	uc := createMockUC(ctrl, uowRepo)
	bookHandler := &BookHandler{uc: uc, observ: observHandler}
	ctx := context.Background()

	res, err := bookHandler.Add(ctx, &pb.BookAddRequest{
		Title:       "Nert",
		Description: "New Desc",
		// Year:        1900,
		Genre: "New Genre",
	})

	assert.Nil(t, res)
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestBook_Add_ErrorUsecase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uowRepo := mocks.NewMockUnitOfWork(ctrl)
	bookMock := mocks.NewMockBookRepository(ctrl)
	bookEventMock := mocks.NewMockBookEventRepository(ctrl)
	observHandler := createMockHandlerObservability(ctrl)
	uc := createMockUC(ctrl, uowRepo)
	bookHandler := &BookHandler{uc: uc, observ: observHandler}
	ctx := context.Background()

	expectedBook := entities.Book{
		Title:       "New Test",
		Description: "New Desc",
		Genre:       "New Genre",
		Year:        1900,
	}
	uowRepo.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(repo *repositories.Repository) error) error {
			bookMock.EXPECT().
				Create(ctx, expectedBook).
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
		
	res, err := bookHandler.Add(ctx, &pb.BookAddRequest{
		Title:       "New Test",
		Description: "New Desc",
		Genre:       "New Genre",
		Year:        1900,
	})

	assert.Nil(t, res)
	assert.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestBook_Add_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uowRepo := mocks.NewMockUnitOfWork(ctrl)
	bookMock := mocks.NewMockBookRepository(ctrl)
	bookEventMock := mocks.NewMockBookEventRepository(ctrl)
	observHandler := createMockHandlerObservability(ctrl)
	uc := createMockUC(ctrl, uowRepo)
	bookHandler := &BookHandler{uc: uc, observ: observHandler}
	ctx := context.Background()

	expectedBook := entities.Book{
		Title:       "New Test",
		Description: "New Desc",
		Genre:       "New Genre",
		Year:        1900,
	}

	uowRepo.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(repo *repositories.Repository) error) error {
			bookMock.EXPECT().
				Create(ctx, expectedBook).
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

	res, err := bookHandler.Add(ctx, &pb.BookAddRequest{
		Title:       "New Test",
		Description: "New Desc",
		Genre:       "New Genre",
		Year:        1900,
	})

	assert.Equal(t, &emptypb.Empty{}, res)
	assert.Nil(t, err)
}
