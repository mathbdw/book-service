package services

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/mathbdw/book/internal/domain/entities"
	errs "github.com/mathbdw/book/internal/errors"
	"github.com/mathbdw/book/mocks"
)

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	op := New(
		mocks.NewMockBookEventRepository(ctrl),
		mocks.NewMockEventPublisher(ctrl),
		mocks.NewMockLogger(ctrl),
	)

	require.NotNil(t, op.eventRepo)
	require.NotNil(t, op.publisher)
}

func setup(t *testing.T) (*gomock.Controller, *mocks.MockBookEventRepository, *mocks.MockEventPublisher, *mocks.MockLogger) {
	ctrl := gomock.NewController(t)
	eventRepo := mocks.NewMockBookEventRepository(ctrl)
	publisher := mocks.NewMockEventPublisher(ctrl)
	logger := mocks.NewMockLogger(ctrl)

	return ctrl, eventRepo, publisher, logger
}
func TestProcessEvent_NoRows(t *testing.T) {
	_, eventRepo, publisher, logger := setup(t)
	op := New(
		eventRepo,
		publisher,
		logger,
	)
	op.batchSize = uint64(50)

	ctx := context.Background()
	eventRepo.EXPECT().
		Lock(ctx, op.batchSize).
		Return(nil, sql.ErrNoRows).
		Times(1)

	//Log - run workers - #N
	logger.EXPECT().Debug(gomock.Any(), gomock.Any()).Times(1)
	//Log - error to repo.lock
	logger.EXPECT().Error(gomock.Any(), gomock.Any()).Times(1)
	op.processEvent(ctx, uint8(1))
}

func TestProcessEvent_NoEvent(t *testing.T) {
	_, eventRepo, publisher, logger := setup(t)
	op := New(
		eventRepo,
		publisher,
		logger,
	)
	op.batchSize = uint64(50)

	ctx := context.Background()
	eventRepo.EXPECT().
		Lock(ctx, op.batchSize).
		Return(nil, errs.ErrNotFound).
		Times(1)

	//Log - run workers - #N
	logger.EXPECT().Debug(gomock.Any(), gomock.Any()).Times(2)

	//Log - events not found
	op.processEvent(ctx, uint8(1))
}

func TestProcessEvent_FalseSend(t *testing.T) {
	_, eventRepo, publisher, logger := setup(t)

	op := New(
		eventRepo,
		publisher,
		logger,
	)
	op.batchSize = uint64(50)

	ctx := context.Background()

	events := []entities.BookEvent{
		{
			ID:      1,
			BookId:  1,
			Payload: []byte("{\"Title\":\"Test\"}"),
			Type:    entities.Created,
		},
	}
	eventRepo.EXPECT().
		Lock(ctx, op.batchSize).
		Return(events, nil).
		Times(1)

	publisher.EXPECT().
		Publish(ctx, &events[0]).
		Return(errors.New("false send")).
		Times(1)

	eventRepo.EXPECT().
		Unlock(ctx, []int64{1}).
		Return(nil).
		Times(1)

	//Log - run workers - #N
	logger.EXPECT().Debug(gomock.Any(), gomock.Any()).Times(1)
	//Log - error to send kafka
	logger.EXPECT().Error(gomock.Any(), gomock.Any()).Times(1)

	op.processEvent(ctx, uint8(1))
}

func TestProcessEvent_FalseUnlock(t *testing.T) {
	_, eventRepo, publisher, logger := setup(t)

	op := New(
		eventRepo,
		publisher,
		logger,
	)
	op.batchSize = uint64(50)

	ctx := context.Background()

	events := []entities.BookEvent{
		{
			ID:      1,
			BookId:  1,
			Payload: []byte("{\"Title\":\"Test\"}"),
			Type:    entities.Created,
		},
	}
	eventRepo.EXPECT().
		Lock(ctx, op.batchSize).
		Return(events, nil).
		Times(1)

	publisher.EXPECT().
		Publish(ctx, &events[0]).
		Return(errors.New("false send")).
		Times(1)

	eventRepo.EXPECT().
		Unlock(ctx, []int64{1}).
		Return(errors.New("error unlock")).
		Times(1)

	//Log - run workers - #N
	logger.EXPECT().Debug(gomock.Any(), gomock.Any()).Times(1)
	//Log - error to send kafka and error eventRepo.unlock
	logger.EXPECT().Error(gomock.Any(), gomock.Any()).Times(2)

	op.processEvent(ctx, uint8(1))
}
