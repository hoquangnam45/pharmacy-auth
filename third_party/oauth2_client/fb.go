package oauth2_client

import (
	"io"
	"net/http"

	"github.com/hoquangnam45/pharmacy-auth/dto"
	"github.com/hoquangnam45/pharmacy-common-go/util"
	h "github.com/hoquangnam45/pharmacy-common-go/util/errorHandler"
)

func FetchFbUserInfo(accessToken string) (*dto.UserInfo, error) {
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
		h.Lift(util.UnmarshalJson(map[string]any{})),
		h.LiftJ(func(userDetail map[string]any) *dto.UserInfo {
			return &dto.UserInfo{
				Id:    userDetail["id"].(string),
				Email: userDetail["email"].(string),
			}
		}),
	).Eval()
}
