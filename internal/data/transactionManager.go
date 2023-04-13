package data

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/hoquangnam45/pharmacy-auth/internal/biz"
	"gorm.io/gorm"
)

type transactionManager struct {
	data *Data
	log  *log.Helper
}

func NewTransactionManager(data *Data, logger log.Logger) biz.TransactionManager {
	return &transactionManager{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (s *transactionManager) Run(f func() error) error {
	return s.data.Transaction(func(*gorm.DB) error {
		return f()
	})
}
