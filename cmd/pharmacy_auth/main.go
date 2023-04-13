package main

import (
	"flag"
	"os"

	"github.com/hoquangnam45/pharmacy-auth/internal/conf"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/env"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
		),
	)
}

func main() {
	flag.Parse()
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)
	c := config.New(
		config.WithSource(
			env.NewSource(""),
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	app, cleanup, err := wireApp(bc.Server, bc.Data, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}

// package main

// import (
// 	"fmt"
// 	"net/http"
// 	"strconv"

// 	"github.com/go-chi/chi/v5"
// 	"github.com/go-chi/chi/v5/middleware"
// 	_ "github.com/golang-migrate/migrate/v4/source/file"
// 	"github.com/hashicorp/consul/api"
// 	"github.com/hoquangnam45/pharmacy-auth/app"
// 	"github.com/hoquangnam45/pharmacy-auth/client"
// 	"github.com/hoquangnam45/pharmacy-auth/controller"
// 	"github.com/hoquangnam45/pharmacy-auth/internal/service"
// 	"github.com/hoquangnam45/pharmacy-common-go/helper/common"
// 	"github.com/hoquangnam45/pharmacy-common-go/microservice/consul"
// 	"github.com/hoquangnam45/pharmacy-common-go/util"
// 	h "github.com/hoquangnam45/pharmacy-common-go/util/errorHandler"
// 	_ "github.com/lib/pq"
// 	"gorm.io/gorm"
// 	"gorm.io/gorm/schema"
// )

// const LISTENNING_PORT = 3001

// func main() {
// 	advertiseIp, advertisePort, clusterPrefix := common.InitializeEcsService(LISTENNING_PORT)
// 	consulClient := common.InitializeConsulClient()

// 	_ = app.NewClusterLoadBalancer(clusterPrefix, consulClient)

// 	h.FlatMap(
// 		h.FactoryE(consulClient.ConnectToConsul),
// 		h.LiftFactoryE[any](func() error {
// 			return consulClient.RegisterService(&api.AgentServiceRegistration{
// 				Name:    util.GetEnvOrDefault("SERVICE_NAME", "pharmacy-auth-svc"),
// 				ID:      advertiseIp + ":" + strconv.Itoa(advertisePort),
// 				Address: advertiseIp,
// 				Port:    advertisePort,
// 				Check: &api.AgentServiceCheck{
// 					Interval: "30s",
// 					Timeout:  "60s",
// 					HTTP:     fmt.Sprintf("http://%s:%d/auth/health", advertiseIp, advertisePort),
// 					Method:   "GET",
// 				}})
// 		}),
// 	).PanicEval()

// 	kvClient := consul.NewKvClient(consulClient)

// 	kvPrefix := util.GetEnvOrDefault("KV_PREFIX", "local")
// 	postgresHost := h.Lift(kvClient.GetKV)(kvPrefix + "/POSTGRES_HOST").PanicEval()
// 	postgresUsername := h.Lift(kvClient.GetKV)(kvPrefix + "/POSTGRES_USERNAME").PanicEval()
// 	postgresPassword := h.Lift(kvClient.GetKV)(kvPrefix + "/POSTGRES_PASSWORD").PanicEval()
// 	postgresDb := h.Lift(kvClient.GetKV)(kvPrefix + "/POSTGRES_DATABASE").PanicEval()
// 	postgresPort := h.FlatMap(
// 		h.Lift(kvClient.GetKV)(kvPrefix+"/POSTGRES_PORT"),
// 		h.Lift(strconv.Atoi),
// 	).DefaultEval(5432)

// 	db := common.InitializePostgresDb(
// 		postgresHost,
// 		postgresUsername,
// 		postgresPassword,
// 		postgresDb,
// 		postgresPort,
// 		&gorm.Config{
// 			NamingStrategy: schema.NamingStrategy{
// 				TablePrefix:   "auth.",
// 				SingularTable: true,
// 			},
// 		},
// 		util.GetEnvOrDefault("MIGRATE_PATH", "./migrations"), 1)

// 	// userInfoClient := client.NewUserInfoClient(lb)
// 	userInfoClientMock := &client.UserInfoMock{}
// 	configService := service.NewConfigService(db)
// 	authService := service.NewAuthService(db, configService, userInfoClientMock)
// 	healthController := controller.NewHealthController()
// 	authController := controller.NewAuthController(authService)

// 	r := chi.NewRouter()
// 	r.Use(middleware.Logger)
// 	r.Get("/auth/health", healthController.HealthCheck)
// 	r.Post("/auth/token", authController.Login)
// 	r.Post("/auth/register", authController.Register)
// 	r.Post("/auth/activate/{userId}", authController.Activate)
// 	r.Post("/auth/check", authController.CheckPermission)
// 	http.ListenAndServe(fmt.Sprintf(":%d", LISTENNING_PORT), r)
// }