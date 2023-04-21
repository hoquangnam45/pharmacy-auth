package server

import (
	v1 "github.com/hoquangnam45/pharmacy-auth/api/auth/v1"
	"github.com/hoquangnam45/pharmacy-auth/internal/conf"
	"github.com/hoquangnam45/pharmacy-auth/internal/service"

	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/hoquangnam45/pharmacy-common-go/util/log"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server, authService *service.Auth, healthCheckService *service.HealthCheckService, logger log.Logger) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
		),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	v1.RegisterAuthServer(srv, authService)
	v1.RegisterHealthCheckServer(srv, healthCheckService)
	return srv
}
