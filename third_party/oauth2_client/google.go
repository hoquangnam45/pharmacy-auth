package oauth2_client

import (
	"context"

	"github.com/hoquangnam45/pharmacy-auth/dto"
	h "github.com/hoquangnam45/pharmacy-common-go/util/errorHandler"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/oauth2/v2"
)

func FetchGoogleUserInfo(accessToken string) (*dto.UserInfo, error) {
	return h.FlatMap3(
		h.FactoryM(func() (*oauth2.Service, error) {
			return oauth2.NewService(context.Background())
		}),
		h.LiftJ(oauth2.NewUserinfoV2Service),
		h.Lift(func(s *oauth2.UserinfoV2Service) (*oauth2.Userinfo, error) {
			return s.Me.Get().Do(googleapi.QueryParameter("access_token", accessToken))
		}),
		h.LiftJ(func(userInfo *oauth2.Userinfo) *dto.UserInfo {
			return &dto.UserInfo{
				Email: userInfo.Email,
				Id:    userInfo.Id,
			}
		}),
	).Eval()
}
