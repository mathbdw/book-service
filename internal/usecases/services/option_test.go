package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWithInterval(t *testing.T) {
	p := &OutboxProcessor{}

	interval := time.Microsecond
	opt := WithInterval(interval)
	opt(p)

	assert.Equal(t, interval, p.interval)
}

func TestWithBatchSize(t *testing.T) {
	p := &OutboxProcessor{}

	batch := uint64(10)
	opt := WithBatchSize(batch)
	opt(p)

	assert.Equal(t, batch, p.batchSize)
}

func TestWithCountWorkers(t *testing.T) {
	p := &OutboxProcessor{}

	count := uint8(5)
	opt := WithCountWorkers(count)
	opt(p)
	assert.Equal(t, count, p.countWorkers)
}
