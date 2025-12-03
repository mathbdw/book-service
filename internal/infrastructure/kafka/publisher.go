package kafka

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/IBM/sarama"

	"github.com/mathbdw/book/internal/domain/entities"
	errs "github.com/mathbdw/book/internal/errors"
	"github.com/mathbdw/book/internal/interfaces/observability"
	"github.com/mathbdw/book/internal/interfaces/publisher"
)

//go:generate mockgen -destination=./../../../../mocks/mock_sync_producer.go -package=mocks github.com/IBM/sarama SyncProducer

type KafkaPublisher struct {
	topic    string
	producer sarama.SyncProducer
	logger   observability.Logger
}

// New - constructor kafka publisher
func New(topic string, producer sarama.SyncProducer, logger observability.Logger) (publisher.EventPublisher, error) {
	if topic == "" {
		return nil, errs.New("publisher.New: topic empty")
	}

	return &KafkaPublisher{topic: topic, producer: producer, logger: logger}, nil
}

func (kp *KafkaPublisher) Publish(ctx context.Context, bookEvent *entities.BookEvent) error {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(bookEvent.Type))

	message := &sarama.ProducerMessage{
		Topic: kp.topic,
		Key:   sarama.StringEncoder(fmt.Sprintf("%d", bookEvent.ID)),
		Value: sarama.ByteEncoder(bookEvent.Payload),
		Headers: []sarama.RecordHeader{
			{Key: []byte("event_type"), Value: buf},
		},
	}

	partition, offset, err := kp.producer.SendMessage(message)

	kp.logger.Debug(
		"publisher.Publish: send",
		map[string]any{
			"partition": partition,
			"offset":    offset,
		},
	)

	return err
}
