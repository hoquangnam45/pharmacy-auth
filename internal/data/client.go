package data

import (
	"github.com/hoquangnam45/pharmacy-auth/internal/model"
	"github.com/hoquangnam45/pharmacy-common-go/util/log"
)

type ClientRepo interface {
	FindByClientID(clientId string) (*model.Client, error)
}

type clientRepo struct {
	data *Data
	log  log.Logger
}

func NewClientRepo(data *Data, logger log.Logger) ClientRepo {
	return &clientRepo{
		data: data,
		log:  logger,
	}
}

func (r *clientRepo) FindByClientID(clientId string) (*model.Client, error) {
	data := &model.Client{}
	err := r.data.Where(&model.Client{ClientId: clientId}).Take(data).Error
	return data, err
}
