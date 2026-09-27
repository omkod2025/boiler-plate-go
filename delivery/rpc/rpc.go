package rpc

import (
	"context"

	"github.com/omkod2025-boop/omgon-notification-service/configs"
	"github.com/omkod2025-boop/omgon-notification-service/pkg/logger"
)

// InitRPCServer เริ่ม RPC server (gRPC/JSON-RPC) — handler ของ RPC ให้วางไว้ใน package นี้
// และเรียกใช้ use case จาก domain/<feature> เช่นเดียวกับ HTTP handler
func InitRPCServer(ctx context.Context, cfg *configs.Config) error {
	// ตรวจสอบ context cancellation
	select {
	case <-ctx.Done():
		logger.Info("Context cancelled, skipping RPC server initialization")
		return nil
	default:
	}

	// TODO: Implement RPC server if needed
	logger.Info("RPC server initialized (placeholder)")
	return nil
}
