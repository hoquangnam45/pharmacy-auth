package data

import (
	"github.com/hoquangnam45/pharmacy-common-go/util/log"
	"gorm.io/gorm"
)

func Query[T any](txManager TransactionManager, queryFn func() (T, error)) (T, error) {
	var ret T
	err := txManager.Run(func() error {
		retI, err := queryFn()
		if err == nil {
			ret = retI
		}
		return err
	})
	return ret, err
}

type TransactionManager interface {
	Run(func() error) error
}

type transactionManager struct {
	data *Data
	log  log.Logger
}

func NewTransactionManager(data *Data, logger log.Logger) TransactionManager {
	return &transactionManager{
		data: data,
		log:  logger,
	}
}

func (s *transactionManager) Run(f func() error) error {
	return s.data.Transaction(func(*gorm.DB) error {
		return f()
	})
}
