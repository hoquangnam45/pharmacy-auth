package data

import (
	"github.com/hoquangnam45/pharmacy-auth/internal/model"

	"github.com/hoquangnam45/pharmacy-common-go/util/log"
)

type RefreshTokenRepo interface {
	Save(*model.RefreshToken) (*model.RefreshToken, error)
	DeleteById(id string) error
	FindById(id string) (*model.RefreshToken, error)
}

type refreshTokenRepo struct {
	data *Data
	log  log.Logger
}

func NewRefreshTokenRepo(data *Data, logger log.Logger) RefreshTokenRepo {
	return &refreshTokenRepo{
		data: data,
		log:  logger,
	}
}

func (r *refreshTokenRepo) Save(g *model.RefreshToken) (*model.RefreshToken, error) {
	if err := r.data.Save(g).Error; err != nil {
		return nil, err
	}
	return g, nil
}

func (r *refreshTokenRepo) DeleteById(id string) error {
	return r.data.Delete(&model.RefreshToken{Id: id}).Error
}

func (r *refreshTokenRepo) FindById(id string) (*model.RefreshToken, error) {
	data := &model.RefreshToken{Id: id}
	err := r.data.Take(data).Error
	return data, err
}
