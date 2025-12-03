package opentelemetry

import (
	"context"
	"testing"

	"github.com/mathbdw/book/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewResource(t *testing.T) {
	cfg := &config.Config{
		Project: config.Project{
			Name:        "test-service",
			Version:     "1.0.0",
			Environment: "testing",
		},
	}

	res, err := NewResource(context.Background(), cfg)

	require.NoError(t, err)
	require.NotNil(t, res)

	attrs := res.Attributes()
	assert.Greater(t, len(attrs), 0, "Resource should have attributes")
}
