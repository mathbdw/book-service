package handlers

import (
	"testing"

	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"

	"github.com/mathbdw/book/internal/usecases/book"
	"github.com/mathbdw/book/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewBookHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRegistrar := mocks.NewMockServiceRegistrar(ctrl)

	mockUowRepo := mocks.NewMockUnitOfWork(ctrl)
	mockBookRepo := mocks.NewMockBookRepository(ctrl)
	observHandler := createMockHandlerObservability(ctrl)

	observUsecase := createMockUsecaseObservability(ctrl)

	addUC := book.NewAddBookUsecase(mockUowRepo, observUsecase)
	getUC := book.NewGetBookUsecase(mockBookRepo, observUsecase)
	listUC := book.NewListBookUsecase(mockBookRepo, observUsecase)
	removeUC := book.NewRemoveBookUsecase(mockUowRepo, observUsecase)

	uc := book.New(
		book.WithAddBookUsecase(addUC),
		book.WithGetBookUsecase(getUC),
		book.WithListBookUsecase(listUC),
		book.WithRemoveBookUsecase(removeUC),
	)

	var registeredService any
	var serviceDesc *grpc.ServiceDesc
	mockRegistrar.EXPECT().
		RegisterService(gomock.Any(), gomock.Any()).
		Do(func(sd *grpc.ServiceDesc, ss any) {
			serviceDesc = sd
			registeredService = ss
		}).
		Times(1)

	NewBookHandler(mockRegistrar, uc, observHandler)

	assert.NotNil(t, registeredService)
	assert.NotNil(t, serviceDesc)

	// Проверяем, что зарегистрирован правильный сервис
	handler, ok := registeredService.(*BookHandler)
	assert.True(t, ok)
	// assert.Equal(t, uc, handler.uc)
	assert.Equal(t, observHandler, handler.observ)

}
