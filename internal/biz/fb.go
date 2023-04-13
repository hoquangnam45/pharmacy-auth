package biz

import (
	"io"
	"net/http"

	"github.com/hoquangnam45/pharmacy-common-go/util"
	h "github.com/hoquangnam45/pharmacy-common-go/util/errorHandler"
)

func FetchFbUserInfo(accessToken string) (*UserInfo, error) {
	return h.FlatMap5(
		h.FactoryM(func() (*http.Request, error) {
			return http.NewRequest("GET", "https://graph.facebook.com/me?fields=id,name,email&access_token="+accessToken, nil)
		}),
		h.Lift(func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		}),
		h.LiftJ(func(res *http.Response) io.ReadCloser {
			return res.Body
		}),
		h.Lift(util.ReadAllThenClose[io.ReadCloser]),
		h.Lift(util.UnmarshalJsonDeref(&map[string]any{})),
		h.LiftJ(func(userDetail map[string]any) *UserInfo {
			return &UserInfo{
				Id:    userDetail["id"].(string),
				Email: userDetail["email"].(string),
			}
		}),
	).Eval()
}
