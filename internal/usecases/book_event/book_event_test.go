package bookevent

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/mathbdw/book/internal/domain/entities"
	errs "github.com/mathbdw/book/internal/errors"
	"github.com/mathbdw/book/mocks"
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

func TestBookEvent_Lock_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bookEventMock := mocks.NewMockBookEventRepository(ctrl)
	observUsecase := createMockUsecaseObservability(ctrl)
	ctx := context.Background()

	bookEventMock.EXPECT().
		Lock(gomock.Any(), uint64(2)).
		Return([]entities.BookEvent{}, errs.New("error"))

	us := NewBookEventUsecase(bookEventMock, observUsecase)

	bookEvents, err := us.Lock(ctx, 2)

	assert.Empty(t, bookEvents)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bookEventUsecase.Lock: set lock events")
	assert.Equal(t, errors.Is(err, errs.New("error")), true)
}

func TestBookEvent_Lock_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bookEventMock := mocks.NewMockBookEventRepository(ctrl)
	observUsecase := createMockUsecaseObservability(ctrl)
	ctx := context.Background()
	events := []entities.BookEvent{
		{ID: 1, BookId: 1, Type: entities.Created, Status: entities.EventStatusLock},
		{ID: 2, BookId: 2, Type: entities.Updated, Status: entities.EventStatusLock},
	}

	bookEventMock.EXPECT().
		Lock(gomock.Any(), uint64(2)).
		Return(events, nil)

	us := NewBookEventUsecase(bookEventMock, observUsecase)

	bookEvents, err := us.Lock(ctx, 2)

	assert.NoError(t, err)
	assert.Equal(t, len(bookEvents), 2)
}

func TestBookEvent_Unlock_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bookEventMock := mocks.NewMockBookEventRepository(ctrl)
	observUsecase := createMockUsecaseObservability(ctrl)
	ctx := context.Background()
	bookEventMock.EXPECT().
		Unlock(gomock.Any(), []int64{1, 2}).
		Return(errs.New("error"))

	uc := NewBookEventUsecase(bookEventMock, observUsecase)
	err := uc.Unlock(ctx, []int64{1, 2})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bookEventUsecase.Unlock: set unlock events")
	assert.True(t, errors.Is(err, errs.New("error")))
}

func TestBookEvent_Unlock_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bookEventMock := mocks.NewMockBookEventRepository(ctrl)
	observUsecase := createMockUsecaseObservability(ctrl)
	ctx := context.Background()
	bookEventMock.EXPECT().
		Unlock(gomock.Any(), []int64{1, 2}).
		Return(nil)

	uc := NewBookEventUsecase(bookEventMock, observUsecase)
	err := uc.Unlock(ctx, []int64{1, 2})

	assert.NoError(t, err)
}

func TestBookEvent_Remove_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bookEventMock := mocks.NewMockBookEventRepository(ctrl)
	observUsecase := createMockUsecaseObservability(ctrl)
	ctx := context.Background()
	bookEventMock.EXPECT().
		Remove(gomock.Any(), []int64{1, 2}).
		Return(errs.New("error"))

	uc := NewBookEventUsecase(bookEventMock, observUsecase)
	err := uc.Remove(ctx, []int64{1, 2})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bookEventUsecase.Remove: delete rows events")
	assert.True(t, errors.Is(err, errs.New("error")))
}

func TestBookEvent_Remove_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bookEventMock := mocks.NewMockBookEventRepository(ctrl)
	observUsecase := createMockUsecaseObservability(ctrl)
	ctx := context.Background()
	bookEventMock.EXPECT().
		Remove(gomock.Any(), []int64{1, 2}).
		Return(nil)

	uc := NewBookEventUsecase(bookEventMock, observUsecase)
	err := uc.Remove(ctx, []int64{1, 2})

	assert.NoError(t, err)
}
