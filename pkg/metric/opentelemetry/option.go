package opentelemetry

import (
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
)

// WithAddress - sets the target endpoint (host and port) the Exporter
func WithAddress(host string, port uint16) otlpmetricgrpc.Option {
	return otlpmetricgrpc.WithEndpoint(fmt.Sprintf("%s:%d", host, port))
}

// WithCompressor - sets the compressor for the gRPC client
func WithCompressor(compressor string) otlpmetricgrpc.Option {
	return otlpmetricgrpc.WithCompressor(compressor)
}

// WithTimeout sets the max amount of time an Exporter will attempt an export
func WithTimeout(duration time.Duration) otlpmetricgrpc.Option {
	return otlpmetricgrpc.WithTimeout(duration)
}

// WithInsecure - disables client transport security for the exporter's gRPC
func WithInsecure(isSet bool) otlpmetricgrpc.Option {
	if isSet {
		return otlpmetricgrpc.WithInsecure()
	}

	return nil
}
