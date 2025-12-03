package book

import (
	"testing"

	"github.com/mathbdw/book/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestWithAddBookUsecase(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockUoWRepo := mocks.NewMockUnitOfWork(ctrl)
	observUsecase := createMockUsecaseObservability(ctrl)
	addUC := NewAddBookUsecase(mockUoWRepo, observUsecase)

	uc := &BookUsecases{}
	opt := WithAddBookUsecase(addUC)
	opt(uc)

	assert.Equal(t, addUC, uc.Add)
}

func TestWithGetBookUsecase(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockBookRepo := mocks.NewMockBookRepository(ctrl)
	observUsecase := createMockUsecaseObservability(ctrl)
	getUC := NewGetBookUsecase(mockBookRepo, observUsecase)

	uc := &BookUsecases{}
	opt := WithGetBookUsecase(getUC)
	opt(uc)

	assert.Equal(t, getUC, uc.Get)
}

func TestWithListBookUsecase(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockBookRepo := mocks.NewMockBookRepository(ctrl)
	observUsecase := createMockUsecaseObservability(ctrl)
	listUC := NewListBookUsecase(mockBookRepo, observUsecase)

	uc := &BookUsecases{}
	opt := WithListBookUsecase(listUC)
	opt(uc)

	assert.Equal(t, listUC, uc.List)
}

func TestWithRemoveBookUsecase(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockUoWRepo := mocks.NewMockUnitOfWork(ctrl)
	observUsecase := createMockUsecaseObservability(ctrl)
	removeUC := NewRemoveBookUsecase(mockUoWRepo, observUsecase)

	uc := &BookUsecases{}
	opt := WithRemoveBookUsecase(removeUC)
	opt(uc)

	assert.Equal(t, removeUC, uc.Remove)
}
