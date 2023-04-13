package server

import (
	v1 "github.com/hoquangnam45/pharmacy-auth/api/auth/v1"
	"github.com/hoquangnam45/pharmacy-auth/internal/conf"
	"github.com/hoquangnam45/pharmacy-auth/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, authService *service.Auth, healthCheckService *service.HealthCheckService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	v1.RegisterAuthHTTPServer(srv, authService)
	v1.RegisterHealthCheckHTTPServer(srv, healthCheckService)
	return srv
}
