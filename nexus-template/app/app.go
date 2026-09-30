// Package app is the service's composition root. Domain routes and consumer
// runners are added here from frozen contracts, never to nexus-kit.
package app

import (
	"context"

	"__MODULE__/configs"
	"github.com/omkod2025/nexus-backend/packages/nexus-kit/service"
)

func Run(ctx context.Context) error {
	return service.Run(ctx, configs.Definition, nil)
}
