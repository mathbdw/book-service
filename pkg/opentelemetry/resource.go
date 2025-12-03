package opentelemetry

import (
	"github.com/mathbdw/book/config"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"golang.org/x/net/context"
)

// NewResource - constructor resource.Resource
func NewResource(ctx context.Context, cfg *config.Config) (*resource.Resource, error) {
	return resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.Project.Name),
			semconv.ServiceVersion(cfg.Project.Version),
			semconv.DeploymentEnvironment(cfg.Project.Environment),
		),
	)
}
