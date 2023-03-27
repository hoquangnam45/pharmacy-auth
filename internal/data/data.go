package data

import (
	"github.com/hoquangnam45/pharmacy-auth/internal/conf"
	"github.com/hoquangnam45/pharmacy-common-go/helper/common"
	"github.com/hoquangnam45/pharmacy-common-go/util"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo)

// Data .
type Data struct {
	*gorm.DB
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	common.InitializePostgresDb(
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
		util.GetEnvOrDefault("MIGRATE_PATH", "./migrations"), 1)
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{}, cleanup, nil
}
