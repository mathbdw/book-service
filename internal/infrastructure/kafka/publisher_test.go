package kafka

import (
	"context"
	"errors"
	"testing"

	"github.com/mathbdw/book/internal/domain/entities"
	"github.com/mathbdw/book/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestPublisher_NewTopicEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	producer := mocks.NewMockSyncProducer(ctrl)
	logger := mocks.NewMockLogger(ctrl)
	kp, err := New("", producer, logger)

	require.Nil(t, kp)
	require.Error(t, err)
	assert.Contains(t, "publisher.New: topic empty", err.Error())
}

func TestPublisher_NewSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	producer := mocks.NewMockSyncProducer(ctrl)
	logger := mocks.NewMockLogger(ctrl)
	kp, err := New("test_topic", producer, logger)

	require.NotNil(t, kp)
	require.NoError(t, err)
}

func TestPulisher_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	producer := mocks.NewMockSyncProducer(ctrl)
	logger := mocks.NewMockLogger(ctrl)
	kp, _ := New("test_topic", producer, logger)
	ctx := context.Background()

	event := &entities.BookEvent{
		ID:      1,
		BookId:  2,
		Type:    entities.Created,
		Payload: []byte{},
	}

	producer.EXPECT().
		SendMessage(gomock.Any()).
		Return(int32(0), int64(0), errors.New("errors send"))

	//Log
	logger.EXPECT().Debug(gomock.Any(), gomock.Any()).Times(1)

	err := kp.Publish(ctx, event)

	require.Error(t, err)
}

func TestPulisher_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	producer := mocks.NewMockSyncProducer(ctrl)
	logger := mocks.NewMockLogger(ctrl)
	kp, _ := New("test_topic", producer, logger)
	ctx := context.Background()

	event := &entities.BookEvent{
		ID:      1,
		BookId:  2,
		Type:    entities.Created,
		Payload: []byte{},
	}

	producer.EXPECT().
		SendMessage(gomock.Any()).
		Return(int32(0), int64(0), errors.New("errors send"))

	//Log
	logger.EXPECT().Debug(gomock.Any(), gomock.Any()).Times(1)

	err := kp.Publish(ctx, event)

	require.Error(t, err)
}
