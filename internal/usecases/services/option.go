package services

import "time"

type Option func(*OutboxProcessor)

// WithInterval - sets interval for run process
func WithInterval(interval time.Duration) Option {
	return func(p *OutboxProcessor) {
		p.interval = interval
	}
}

// WithBatchsize - sets the sample size for the book_event
func WithBatchSize(batch uint64) Option {
	return func(p *OutboxProcessor) {
		p.batchSize = batch
	}
}

// WithCountWorkers - sets the number of workers
func WithCountWorkers(count uint8) Option {
	return func(p *OutboxProcessor) {
		p.countWorkers = count
	}
}
