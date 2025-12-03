package producer

import (
	"github.com/IBM/sarama"
)

type Service struct {
	config  *sarama.Config
	brokers []string
}

// New - constructor Service producer
func New(opts ...Option) *Service {
	s := &Service{
		config: sarama.NewConfig(),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Service) Start() (sarama.SyncProducer, error) {
	return sarama.NewSyncProducer(s.brokers, s.config)
}
