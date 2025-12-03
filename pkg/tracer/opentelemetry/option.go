package opentelemetry

import (
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
)

// WithAddress - sets the target endpoint (host and port) the Exporter
func WithAddress(host string, port uint16) otlptracegrpc.Option {
	return otlptracegrpc.WithEndpoint(fmt.Sprintf("%s:%d", host, port))
}

// WithCompressor - sets the compressor for the gRPC client
func WithCompressor(compressor string) otlptracegrpc.Option {
	return otlptracegrpc.WithCompressor(compressor)
}

// WithTimeout sets the max amount of time a client will attempt to export a
// batch of spans.
func WithTimeout(duration time.Duration) otlptracegrpc.Option {
	return otlptracegrpc.WithTimeout(duration)
}

// WithInsecure - disables client transport security for the exporter's gRPC
func WithInsecure(isSet bool) otlptracegrpc.Option {
	if isSet {
		return otlptracegrpc.WithInsecure()
	}

	return nil
}
