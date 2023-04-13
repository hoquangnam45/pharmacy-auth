package data

import (
	"database/sql"

	"github.com/hoquangnam45/pharmacy-auth/internal/conf"
	"github.com/hoquangnam45/pharmacy-common-go/helper/common"
	h "github.com/hoquangnam45/pharmacy-common-go/util/errorHandler"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewLoginDetailRepo, NewRefreshTokenRepo, NewTransactionManager, NewClientRepo)

// Data .
type Data struct {
	*gorm.DB
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	dbConfig := c.Database
	db := common.InitializePostgresDb(
		dbConfig.Host,
		dbConfig.Username,
		dbConfig.Password,
		dbConfig.Database,
		int(dbConfig.Port),
		&gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				TablePrefix:   "auth.",
				SingularTable: true,
			},
		},
		dbConfig.MigratePath,
		1)
	cleanup := func() {
		h.FlatMap(
			h.FactoryM(db.DB),
			h.PeekE(func(con *sql.DB) error {
				return con.Close()
			}),
		).PanicEval()
	}
	return &Data{db}, cleanup, nil
}