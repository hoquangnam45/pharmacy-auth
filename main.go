package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/hashicorp/consul/api"
	"github.com/hoquangnam45/pharmacy-auth/app"
	"github.com/hoquangnam45/pharmacy-auth/client"
	"github.com/hoquangnam45/pharmacy-auth/controller"
	"github.com/hoquangnam45/pharmacy-auth/service"
	"github.com/hoquangnam45/pharmacy-common-go/helper/common"
	"github.com/hoquangnam45/pharmacy-common-go/helper/env"
	h "github.com/hoquangnam45/pharmacy-common-go/helper/errorHandler"
	"github.com/hoquangnam45/pharmacy-common-go/microservice/consul"
	_ "github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

const LISTENNING_PORT = 3001
const HEALTH_CHECK_PATH = "/health"

func main() {
	advertiseIp, advertisePort, clusterPrefix := common.InitializeEcsService(LISTENNING_PORT)
	consulClient := common.InitializeConsulClient()

	_ = app.NewClusterLoadBalancer(clusterPrefix, consulClient)

	h.FlatMap(
		h.FactoryE(consulClient.ConnectToConsul),
		h.LiftFactoryE[any](func() error {
			return consulClient.RegisterService(&api.AgentServiceRegistration{
				Name:    env.GetEnvOrDefault("SERVICE_NAME", "pharmacy-auth-svc"),
				ID:      advertiseIp + ":" + strconv.Itoa(advertisePort),
				Address: advertiseIp,
				Port:    advertisePort,
				Check: &api.AgentServiceCheck{
					Interval: "30s",
					Timeout:  "60s",
					HTTP:     fmt.Sprintf("http://%s:%d/health", advertiseIp, advertisePort),
					Method:   "GET",
				}})
		}),
	).PanicEval()

	kvClient := consul.NewKvClient(consulClient)

	kvPrefix := env.GetEnvOrDefault("KV_PREFIX", "local")
	postgresHost := h.Lift(kvClient.GetKV)(kvPrefix + "/POSTGRES_HOST").PanicEval()
	postgresUsername := h.Lift(kvClient.GetKV)(kvPrefix + "/POSTGRES_USERNAME").PanicEval()
	postgresPassword := h.Lift(kvClient.GetKV)(kvPrefix + "/POSTGRES_PASSWORD").PanicEval()
	postgresDb := h.Lift(kvClient.GetKV)(kvPrefix + "/POSTGRES_DATABASE").PanicEval()
	postgresPort := h.FlatMap(
		h.Lift(kvClient.GetKV)(kvPrefix+"/POSTGRES_PORT"),
		h.Lift(strconv.Atoi),
	).DefaultEval(5432)

	db := common.InitializePostgresDb(
		postgresHost,
		postgresUsername,
		postgresPassword,
		postgresDb,
		postgresPort,
		&gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				TablePrefix:   "auth.",
				SingularTable: true,
			},
		},
		env.GetEnvOrDefault("MIGRATE_PATH", "./migrations"), 1)

	// userInfoClient := client.NewUserInfoClient(lb)
	userInfoClientMock := &client.UserInfoMock{}
	configService := service.NewConfigService(db)
	authService := service.NewAuthService(db, configService, userInfoClientMock)
	healthController := controller.NewHealthController()
	authController := controller.NewAuthController(authService)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get(HEALTH_CHECK_PATH, healthController.HealthCheck)
	r.Post("/auth/token", authController.Login)
	r.Post("/auth/register", authController.Register)
	r.Post("/auth/activate/{userId}", authController.Activate)
	r.Post("/auth/check", authController.CheckPermission)
	http.ListenAndServe(fmt.Sprintf(":%d", LISTENNING_PORT), r)
}
