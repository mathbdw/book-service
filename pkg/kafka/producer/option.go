package producer

import "github.com/IBM/sarama"

// Option -.
type Option func(*Service)

// WithReturnSuccesses - sets slice brokers
func WithBrokers(brokers []string) Option {
	return func(s *Service) {
		s.brokers = brokers
	}
}

// WithReturnSuccesses - sets the flag for returning delivered messages via the successful action channel
func WithReturnSuccesses(flag bool) Option {
	return func(s *Service) {
		s.config.Producer.Return.Successes = flag
	}
}

// WithRequiredAcks - sets the level of acknowledgement reliability needed from the broker
func WithRequiredAcks(ack int16) Option {
	return func(s *Service) {
		s.config.Producer.RequiredAcks = sarama.RequiredAcks(ack)
	}
}

// WithCompression - sets the type of compression to use on messages
func WithCompression(copression int8) Option {
	return func(s *Service) {
		s.config.Producer.Compression = sarama.CompressionCodec(copression)
	}
}

// WithCompression - sets PartitionerConstructor
func WithPartitioner(partioner string) Option {
	return func(s *Service) {
		s.config.Producer.Partitioner = convertPartitioner(partioner)
	}
}

// convertPartitioner - .
func convertPartitioner(partitioner string) sarama.PartitionerConstructor {
	switch partitioner {
	case "random":
		return sarama.NewRandomPartitioner
	default:
		return sarama.NewHashPartitioner
	}
}
