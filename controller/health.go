package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/hellofresh/health-go/v5"
)

type HealthController struct{}

var healthCheck http.HandlerFunc

func NewHealthController() *HealthController {
	return &HealthController{}
}

func (c *HealthController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	healthCheck(w, r)
}

func init() {
	h, err := health.New(health.WithComponent(health.Component{
		Name:    "pharmacy-auth",
		Version: "0.0.1-SNAPSHOT",
	}), health.WithChecks(health.Config{
		Name:      "ping",
		Timeout:   time.Second,
		SkipOnErr: true,
		Check: func(ctx context.Context) error {
			return nil
		}},
	))
	if err != nil {
		panic(err)
	}
	healthCheck = h.HandlerFunc
}
