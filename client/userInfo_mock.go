package client

import (
	"github.com/google/uuid"
	"github.com/hoquangnam45/pharmacy-auth/dto"
	"github.com/hoquangnam45/pharmacy-common-go/util/request"
)

type UserInfoMock struct{}

func (s *UserInfoMock) FetchUserInfo(username string, email string, phoneNumber string) (*dto.UserInfo, error) {
	// Already existed user
	if email == "hoquangnam45@gmail.com" {
		return &dto.UserInfo{
			Id:          "00000000-0000-0000-0000-000000000000",
			Email:       "hoquangnam45@gmail.com",
			Username:    "hoquangnam45",
			PhoneNumber: "+840912345678",
		}, nil
	} else {
		return nil, request.NewErrorResponse("not found", 404)
	}
}

func (s *UserInfoMock) CreateUserInfo(username, email, phoneNumber string) (*dto.UserInfo, error) {
	// New user
	if email == "hoquangnam46@gmail.com" {
		return &dto.UserInfo{
			Id:          uuid.New().String(),
			Email:       uuid.New().String(),
			Username:    uuid.New().String(),
			PhoneNumber: uuid.New().String(),
		}, nil
	}
	// Already existed user
	if email == "hoquangnam45@gmail.com" {
		return nil, request.NewErrorResponse("credentials already exist", 409)
	}
	return nil, request.NewErrorResponse("bad request", 400)
}
