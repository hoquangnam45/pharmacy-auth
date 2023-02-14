package client

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/hoquangnam45/pharmacy-auth/app"
	"github.com/hoquangnam45/pharmacy-auth/dto"
	h "github.com/hoquangnam45/pharmacy-common-go/helper/errorHandler"

	"github.com/hoquangnam45/pharmacy-common-go/util/request"
)

var ErrUserNotExist = errors.New("user not exist")

type IUserInfo interface {
	FetchUserInfo(username, email, phoneNumber string) (*dto.UserInfo, error)
	CreateUserInfo(username, email, phoneNumber string) (*dto.UserInfo, error)
}

type UserInfo struct {
	lb *app.LoadBalancer
}

func NewUserInfoClient(lb *app.LoadBalancer) *UserInfo {
	return &UserInfo{lb}
}

func (s *UserInfo) FetchUserInfo(username, email, phoneNumber string) (*dto.UserInfo, error) {
	return h.FlatMap3(
		h.Lift(s.lb.LoadBalancing)("pharmacy-user-svc"),
		h.Lift(func(addr string) (string, error) {
			query := ""
			if email != "" {
				query = fmt.Sprintf("email=%s", email)
			} else if phoneNumber != "" {
				query = fmt.Sprintf("phoneNumber=%s", email)
			} else if username != "" {
				query = fmt.Sprintf("username=%s", email)
			}
			return fmt.Sprintf("%s/user?%s", addr, query), nil
		}),
		h.Lift(http.Get),
		h.Lift(func(r *http.Response) (*dto.UserInfo, error) {
			defer r.Body.Close()
			return h.FlatMap(
				h.Lift(request.NewRequestResponse)(r),
				h.Lift(func(resp *request.Response) (*dto.UserInfo, error) {
					switch resp.StatusCode {
					case 200:
						return request.ToJsonResponse[*dto.UserInfo](resp).Get(&dto.UserInfo{})
					case 404:
						return nil, ErrUserNotExist
					default:
						return nil, request.ToErrorResponse(resp)
					}
				})).Eval()
		})).Eval()
}

func (s *UserInfo) CreateUserInfo(username, email, phoneNumber string) (*dto.UserInfo, error) {
	return h.FlatMap3(
		h.Lift(s.lb.LoadBalancing)("pharmacy-user-svc"),
		h.Lift(func(addr string) (string, error) {
			query := ""
			if email != "" {
				query = fmt.Sprintf("email=%s", email)
			} else if phoneNumber != "" {
				query = fmt.Sprintf("phoneNumber=%s", email)
			} else if username != "" {
				query = fmt.Sprintf("username=%s", email)
			}
			return fmt.Sprintf("%s/user?%s", addr, query), nil
		}),
		h.Lift(http.Get),
		h.Lift(func(r *http.Response) (*dto.UserInfo, error) {
			defer r.Body.Close()
			return h.FlatMap(
				h.Lift(request.NewRequestResponse)(r),
				h.Lift(func(resp *request.Response) (*dto.UserInfo, error) {
					switch resp.StatusCode {
					case 200:
						return request.ToJsonResponse[*dto.UserInfo](resp).Get(&dto.UserInfo{})
					default:
						return nil, request.ToErrorResponse(resp)
					}
				})).Eval()
		})).Eval()
}
