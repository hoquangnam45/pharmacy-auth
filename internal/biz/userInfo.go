package biz

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/hoquangnam45/pharmacy-auth/internal/util"
	h "github.com/hoquangnam45/pharmacy-common-go/util/errorHandler"

	"github.com/hoquangnam45/pharmacy-common-go/util/request"
)

var ErrUserNotExist = errors.New("user not exist")

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
	lb *util.LoadBalancer
}

func NewUserInfoClient(lb *util.LoadBalancer) UserInfoClient {
	return &userInfoClient{lb}
}

func (s *userInfoClient) FetchUserInfo(username, email, phoneNumber *string) (*UserInfo, error) {
	return h.FlatMap3(
		h.Lift(s.lb.LoadBalancing)("pharmacy-user-svc"),
		h.Lift(func(addr string) (string, error) {
			query := ""
			if email != nil {
				query = fmt.Sprintf("email=%s", email)
			} else if phoneNumber != nil {
				query = fmt.Sprintf("phoneNumber=%s", email)
			} else if username != nil {
				query = fmt.Sprintf("username=%s", email)
			}
			return fmt.Sprintf("%s/user?%s", addr, query), nil
		}),
		h.Lift(http.Get),
		h.Lift(func(r *http.Response) (*UserInfo, error) {
			defer r.Body.Close()
			return h.FlatMap(
				h.Lift(request.NewRequestResponse)(r),
				h.Lift(func(resp *request.Response) (*UserInfo, error) {
					switch resp.StatusCode {
					case 200:
						return request.ToJsonResponse[UserInfo](resp).Get(&UserInfo{})
					case 404:
						return nil, ErrUserNotExist
					default:
						return nil, request.ToErrorResponse(resp)
					}
				})).Eval()
		})).Eval()
}

func (s *userInfoClient) CreateUserInfo(username, email, phoneNumber *string) (*UserInfo, error) {
	return h.FlatMap3(
		h.Lift(s.lb.LoadBalancing)("pharmacy-user-svc"),
		h.Lift(func(addr string) (string, error) {
			query := ""
			if email != nil {
				query = fmt.Sprintf("email=%s", email)
			} else if phoneNumber != nil {
				query = fmt.Sprintf("phoneNumber=%s", email)
			} else if username != nil {
				query = fmt.Sprintf("username=%s", email)
			}
			return fmt.Sprintf("%s/user?%s", addr, query), nil
		}),
		h.Lift(http.Get),
		h.Lift(func(r *http.Response) (*UserInfo, error) {
			defer r.Body.Close()
			return h.FlatMap(
				h.Lift(request.NewRequestResponse)(r),
				h.Lift(func(resp *request.Response) (*UserInfo, error) {
					switch resp.StatusCode {
					case 200:
						return request.ToJsonResponse[UserInfo](resp).Get(&UserInfo{})
					default:
						return nil, request.ToErrorResponse(resp)
					}
				})).Eval()
		})).Eval()
}

// TODO: Implement this
func (s *userInfoClient) RemoveUserInfo(username, email, phoneNumber *string) error {
	return nil
}
