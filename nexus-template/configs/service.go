package configs

import "github.com/omkod2025/nexus-backend/packages/nexus-kit/service"

var Definition = service.Definition{
	Name:        "__SERVICE__",
	Roles:       []string{__ROLES__},
	DefaultRole: "__DEFAULT_ROLE__",
	Port:        __PORT__,
	HasDB:       __HAS_DB__,
	HasBroker:   __HAS_BROKER__,
}
