package opentelemetry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
)

func TestWithAddress(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		port     uint16
		expected string
	}{
		{
			name:     "localhost with default port",
			host:     "localhost",
			port:     4317,
			expected: "localhost:4317",
		},
		{
			name:     "IPv6 address",
			host:     "::1",
			port:     4317,
			expected: "::1:4317",
		},
		{
			name:     "empty host",
			host:     "",
			port:     4317,
			expected: ":4317",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			option := WithAddress(tt.host, tt.port)

			require.NotNil(t, option)
			assert.Implements(t, (*otlptracegrpc.Option)(nil), option)
		})
	}
}

func TestWithCompressor(t *testing.T) {
	option := WithCompressor("gzip")

	require.NotNil(t, option)
	assert.Implements(t, (*otlptracegrpc.Option)(nil), option)
}

func TestWithTimeout(t *testing.T) {
	option := WithTimeout(time.Second)

	require.NotNil(t, option)
	assert.Implements(t, (*otlptracegrpc.Option)(nil), option)
}

func TestWithInsecure_Without(t *testing.T) {
	option := WithInsecure(false)

	require.Nil(t, option)
}

func TestWithInsecure_With(t *testing.T) {
	option := WithInsecure(true)

	require.NotNil(t, option)
	assert.Implements(t, (*otlptracegrpc.Option)(nil), option)
}
