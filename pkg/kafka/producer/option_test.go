package producer

import (
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"
)

func TestWithBrokers(t *testing.T) {
	expected := []string{"localhost1"}

	s := &Service{}
	opt := WithBrokers(expected)
	opt(s)

	require.Equal(t, expected, s.brokers)
}

func TestWithReturnSuccesses(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
		flag     bool
	}{
		{"True", true, true},
		{"False", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{config: sarama.NewConfig()}
			opt := WithReturnSuccesses(tt.flag)
			opt(s)

			require.Equal(t, tt.expected, s.config.Producer.Return.Successes)
		})
	}
}

func TestWithRequiredAcks(t *testing.T) {
	s := &Service{config: sarama.NewConfig()}
	opt := WithRequiredAcks(1)
	opt(s)

	require.Equal(t, sarama.WaitForLocal, s.config.Producer.RequiredAcks)
}

func TestWithCompression(t *testing.T) {
	expected := sarama.CompressionSnappy
	s := &Service{config: sarama.NewConfig()}
	opt := WithCompression(2)
	opt(s)

	require.Equal(t, expected, s.config.Producer.Compression)
}