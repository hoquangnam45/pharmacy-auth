package biz

import (
	"context"

	"github.com/hoquangnam45/pharmacy-common-go/microservice/consul"
	"github.com/hoquangnam45/pharmacy-common-go/util"
	h "github.com/hoquangnam45/pharmacy-common-go/util/errorHandler"
)

type UserInfoClient interface {
	FetchUserInfo(username, email, phoneNumber *string) (*UserInfo, error)
	CreateUserInfo(username, email, phoneNumber *string) (*UserInfo, error)
	RemoveUserInfo(username, email, phoneNumber *string) error
}

type UserInfo struct {
	Id          string
	Email       string
	Username    string
	PhoneNumber string
}

type userInfoClient struct {
	client *consul.Client
}

func NewUserInfoClient(c *consul.Client) UserInfoClient {
	return &userInfoClient{
		client: c,
	}
}

func (s *userInfoClient) FetchUserInfo(username, email, phoneNumber *string) (*UserInfo, error) {
	return h.FlatMap(
		h.FactoryM(func() (map[string]any, error) {
			resp := map[string]any{}
			err := s.client.CallService(context.Background(), "user", "FetchUserInfo", "", map[string]*string{
				"email":       email,
				"phoneNumber": phoneNumber,
				"username":    username,
			}, map[string]any{})
			return resp, err
		}),
		h.Lift(MapUserInfo)).Eval()
}

func (s *userInfoClient) CreateUserInfo(username, email, phoneNumber *string) (*UserInfo, error) {
	return h.FlatMap(
		h.FactoryM(func() (map[string]any, error) {
			resp := map[string]any{}
			err := s.client.CallService(context.Background(), "user", "CreateUserInfo", "", map[string]*string{
				"email":       email,
				"phoneNumber": phoneNumber,
				"username":    username,
			}, map[string]any{})
			return resp, err
		}),
		h.Lift(MapUserInfo)).Eval()
}

func (s *userInfoClient) RemoveUserInfo(username, email, phoneNumber *string) error {
	return h.FactoryM(func() (map[string]any, error) {
		resp := map[string]any{}

		err := s.client.CallService(context.Background(), "user", "CreateUserInfo", "", map[string]*string{
			"email":       email,
			"phoneNumber": phoneNumber,
			"username":    username,
		}, map[string]any{})
		return resp, err
	}).Error()
}

func MapUserInfo(m map[string]any) (*UserInfo, error) {
	return h.FlatMap(
		h.Lift(util.MarshalJson[map[string]any])(m),
		h.Lift(util.UnmarshalJson(&UserInfo{})),
	).Eval()
}
