package biz

import (
	"github.com/google/uuid"
	"github.com/hoquangnam45/pharmacy-common-go/util/request"
)

func NewUserInfoClientMock() UserInfoClient {
	return &userInfoMock{}
}

type userInfoMock struct{}

func (s *userInfoMock) FetchUserInfo(username *string, email *string, phoneNumber *string) (*UserInfo, error) {
	// Already existed user
	if email != nil && *email == "hoquangnam45@gmail.com" {
		return &UserInfo{
			Id:          "00000000-0000-0000-0000-000000000000",
			Email:       "hoquangnam45@gmail.com",
			Username:    "hoquangnam45",
			PhoneNumber: "+840912345678",
		}, nil
	} else {
		return nil, request.NewErrorResponse("not found", 404)
	}
}

func (s *userInfoMock) CreateUserInfo(username, email, phoneNumber *string) (*UserInfo, error) {
	// New user
	if email != nil && *email == "hoquangnam46@gmail.com" {
		return &UserInfo{
			Id:          uuid.New().String(),
			Email:       uuid.New().String(),
			Username:    uuid.New().String(),
			PhoneNumber: uuid.New().String(),
		}, nil
	}
	// Already existed user
	if email != nil && *email == "hoquangnam45@gmail.com" {
		return nil, request.NewErrorResponse("credentials already exist", 409)
	}
	return nil, request.NewErrorResponse("bad request", 400)
}

// TODO: Implement this
func (s *userInfoMock) RemoveUserInfo(username, email, phoneNumber *string) error {
	return nil
}
