package data

import (
	"github.com/google/uuid"
	"github.com/hoquangnam45/pharmacy-auth/internal/model"
	"github.com/hoquangnam45/pharmacy-common-go/util/log"
)

type LoginDetailRepo interface {
	Save(*model.LoginDetail) (*model.LoginDetail, error)
	FindByID(uuid.UUID) (*model.LoginDetail, error)
	FindByUserId(userId string) (*model.LoginDetail, error)
}

type loginDetailRepo struct {
	data *Data
	log  log.Logger
}

func NewLoginDetailRepo(data *Data, logger log.Logger) LoginDetailRepo {
	return &loginDetailRepo{
		data: data,
		log:  logger,
	}
}

func (r *loginDetailRepo) Save(g *model.LoginDetail) (*model.LoginDetail, error) {
	if err := r.data.Save(g).Error; err != nil {
		return nil, err
	}
	return g, nil
}

func (r *loginDetailRepo) FindByID(id uuid.UUID) (*model.LoginDetail, error) {
	loginDetail := &model.LoginDetail{}
	err := r.data.Where(&model.LoginDetail{Id: id}).Take(loginDetail).Error
	if err != nil {
		return nil, err
	}
	return loginDetail, nil
}

func (r *loginDetailRepo) FindByUserId(userId string) (*model.LoginDetail, error) {
	loginDetail := &model.LoginDetail{}
	err := r.data.Where(&model.LoginDetail{UserId: userId}).Take(loginDetail).Error
	if err != nil {
		return nil, err
	}
	return loginDetail, nil
}
