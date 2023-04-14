package data

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/hoquangnam45/pharmacy-auth/internal/biz"
)

type clientRepo struct {
	data *Data
	log  *log.Helper
}

func NewClientRepo(data *Data, logger log.Logger) biz.ClientRepo {
	return &clientRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *clientRepo) FindByClientID(clientId string) (*biz.Client, error) {
	data := &biz.Client{}
	err := r.data.Where(&biz.Client{ClientId: clientId}).Take(data).Error
	return data, err
}
