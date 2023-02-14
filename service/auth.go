package service

import (
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hoquangnam45/pharmacy-auth/client"
	"github.com/hoquangnam45/pharmacy-auth/constant/grantType"
	"github.com/hoquangnam45/pharmacy-auth/constant/oauthProviderType"
	"github.com/hoquangnam45/pharmacy-auth/dto"
	"github.com/hoquangnam45/pharmacy-auth/model"
	"github.com/hoquangnam45/pharmacy-auth/service/oauth2_client"
	"github.com/hoquangnam45/pharmacy-auth/utils"
	"github.com/hoquangnam45/pharmacy-common-go/helper/db"
	h "github.com/hoquangnam45/pharmacy-common-go/helper/errorHandler"
	"github.com/hoquangnam45/pharmacy-common-go/util"
	"github.com/hoquangnam45/pharmacy-common-go/util/request"
	"github.com/hoquangnam45/pharmacy-common-go/util/response"
	"gorm.io/gorm"
)

type Auth struct {
	db             *gorm.DB
	configService  *Config
	userInfoClient client.IUserInfo
}

type TrustedTpInfoFetcher func(accessToken string) (*dto.UserInfo, error)

var (
	ErrNotSupportOauthProvider = response.NewResponseError(401, "not supported openid provider")
	ErrNotSupportGrantType     = response.NewResponseError(401, "not supported grant type")
	ErrInvalidTpAccessToken    = response.NewResponseError(401, "invalid trusted third party access token")
	ErrInvalidCredential       = response.NewResponseError(401, "invalid credential")
	ErrCredentialAlreadyExist  = response.NewResponseError(409, "crendential already exist")
	ErrInvalidClientId         = response.NewResponseError(401, "invalid client id")
	ErrUnauthorizedAccess      = response.NewResponseError(401, "unauthorized access")
)

func NewAuthService(db *gorm.DB, configService *Config, userInfoClient client.IUserInfo) *Auth {
	return &Auth{db, configService, userInfoClient}
}

func (s *Auth) Login(grantRequest *dto.GrantRequest) (*dto.Authentication, error) {
	switch grantRequest.GrantType.Normalize() {
	case grantType.Password:
		return s.passwordAuthenticated(grantRequest)
	case grantType.RefreshToken:
		return s.refreshTokenAuthenticated(grantRequest)
	case grantType.TrustedTp:
		switch grantRequest.Provider.Normalize() {
		case oauthProviderType.Facebook:
			return s.trustedTpAuthenticated(oauth2_client.FetchFbUserInfo, grantRequest.AccessToken, grantRequest.ClientId)
		case oauthProviderType.Google:
			return s.trustedTpAuthenticated(oauth2_client.FetchGoogleUserInfo, grantRequest.AccessToken, grantRequest.ClientId)
		default:
			return nil, fmt.Errorf("%w %s", ErrNotSupportOauthProvider, grantRequest.Provider)
		}
	default:
		return nil, fmt.Errorf("%w %s", ErrNotSupportGrantType, grantRequest.GrantType)
	}
}

func (s *Auth) Register(registerRequest *dto.GrantRequest) (*dto.Authentication, error) {
	return h.FlatMap2(
		h.FactoryM(func() (*dto.UserInfo, error) {
			return s.userInfoClient.CreateUserInfo(registerRequest.Username, registerRequest.Email, registerRequest.PhoneNumber)
		}),
		h.PeekE(func(userInfo *dto.UserInfo) error {
			return h.FlatMap(
				h.Lift(util.HashPassword)(registerRequest.Password),
				h.Lift(func(pass string) (*model.LoginDetail, error) {
					loginDetail := &model.LoginDetail{
						UserId:    userInfo.Id,
						Password:  pass,
						Activated: false,
					}
					if err := s.db.Create(loginDetail).Error; err != nil {
						return nil, err
					}
					return loginDetail, nil
				})).Error()
		}),
		h.Lift(func(userInfo *dto.UserInfo) (*dto.Authentication, error) {
			hashedPassword, err := util.HashPassword(registerRequest.Password)
			if err != nil {
				return nil, err
			}
			if err := s.db.Create(&model.LoginDetail{
				UserId:    userInfo.Id,
				Password:  hashedPassword,
				Activated: false}).Error; err != nil {
				return nil, err
			}
			return &dto.Authentication{Subject: userInfo.Id, Authenticated: true}, nil
		}),
	).EvalWithHandlerE(func(err error) error {
		requestErr := &request.Error{}
		if db.IsDuplicatedError(err) || errors.As(err, &requestErr) && requestErr.StatusCode == 409 {
			return ErrCredentialAlreadyExist
		}
		return nil
	})
}

func (s *Auth) GrantAccess(authentication *dto.Authentication) (*dto.GrantAccess, error) {
	if authentication == nil || !authentication.Authenticated {
		return nil, ErrUnauthorizedAccess
	}
	client := &model.Client{}
	if err := s.db.Where(&model.Client{ClientId: authentication.ClientId}).First(client).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInvalidClientId
	} else if err != nil {
		return nil, err
	} else if !client.Active {
		return nil, ErrUnauthorizedAccess
	}
	switch authentication.GrantType.Normalize() {
	case grantType.Password:
		fallthrough
	case grantType.TrustedTp:
		return s.grantAccessByPassword(authentication, client)
	case grantType.RefreshToken:
		return s.grantAccessByRefreshToken(authentication, client)
	default:
		return nil, ErrUnauthorizedAccess
	}
}

func (s *Auth) grantAccessByRefreshToken(authentication *dto.Authentication, client *model.Client) (*dto.GrantAccess, error) {
	refreshToken := authentication.Credential.(*model.RefreshToken)
	grantAccess := &dto.GrantAccess{}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		if err := tx.Where(refreshToken).Delete(&model.RefreshToken{}).Error; err != nil {
			return err
		}
		newRefreshToken := &model.RefreshToken{
			Id:        base64.StdEncoding.EncodeToString([]byte(uuid.New().String())),
			IssuedAt:  now,
			ExpiredAt: now.Add(client.RefreshTokenTtl),
			Client:    client,
			Subject:   authentication.Subject,
		}
		if err := tx.Create(newRefreshToken).Error; err != nil {
			return err
		}
		accessToken, err := utils.GenerateAccessToken(newRefreshToken, client)
		if err != nil {
			return err
		}
		grantAccess = &dto.GrantAccess{
			RefreshToken: newRefreshToken.Id,
			AccessToken:  accessToken,
			Subject:      newRefreshToken.Subject,
			IssuedAt:     newRefreshToken.ExpiredAt,
			ExpiredAt:    newRefreshToken.ExpiredAt,
			ExpiredIn:    client.AccessTokenTtl,
			ClientId:     client.ClientId,
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return grantAccess, nil
}

func (s *Auth) grantAccessByPassword(authentication *dto.Authentication, client *model.Client) (*dto.GrantAccess, error) {
	now := time.Now()
	newRefreshToken := &model.RefreshToken{
		Id:        base64.StdEncoding.EncodeToString([]byte(uuid.New().String())),
		IssuedAt:  now,
		ExpiredAt: now.Add(client.RefreshTokenTtl),
		Client:    client,
		Subject:   authentication.Subject,
	}
	accessToken, err := utils.GenerateAccessToken(newRefreshToken, client)
	if err != nil {
		return nil, err
	}
	grantAccess := &dto.GrantAccess{
		RefreshToken: newRefreshToken.Id,
		AccessToken:  accessToken,
		Subject:      newRefreshToken.Subject,
		IssuedAt:     newRefreshToken.ExpiredAt,
		ExpiredAt:    newRefreshToken.ExpiredAt,
		ExpiredIn:    client.AccessTokenTtl,
		ClientId:     client.ClientId,
	}
	if err := s.db.Create(newRefreshToken).Error; err != nil {
		return nil, err
	}
	return grantAccess, nil
}

func (s *Auth) trustedTpAuthenticated(fetcher TrustedTpInfoFetcher, accessToken string, clientId string) (*dto.Authentication, error) {
	return h.FlatMap2(
		h.Lift(fetcher)(accessToken),
		h.Lift(func(userInfo *dto.UserInfo) (*dto.UserInfo, error) {
			if userInfo, err := s.userInfoClient.FetchUserInfo("", userInfo.Email, ""); errors.Is(err, client.ErrUserNotExist) {
				return s.userInfoClient.CreateUserInfo("", userInfo.Email, "")
			} else {
				return userInfo, nil
			}
		}),
		h.Lift(func(userInfo *dto.UserInfo) (*dto.Authentication, error) {
			return &dto.Authentication{
				Subject:       userInfo.Id,
				Authenticated: true,
				Credential:    accessToken,
				GrantType:     grantType.TrustedTp,
				ClientId:      clientId,
			}, nil
		}),
	).EvalWithHandlerE(func(err error) error {
		requestErr := &request.Error{}
		if errors.As(err, &requestErr) || errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvalidCredential
		}
		return nil
	})
}

func (s *Auth) passwordAuthenticated(loginRequest *dto.GrantRequest) (*dto.Authentication, error) {
	return h.FlatMap(
		h.FactoryM(func() (*dto.UserInfo, error) {
			return s.userInfoClient.FetchUserInfo(loginRequest.Username, loginRequest.Email, loginRequest.PhoneNumber)
		}),
		h.Lift(func(userInfo *dto.UserInfo) (*dto.Authentication, error) {
			loginDetail := model.LoginDetail{}
			if err := s.db.Where(&model.LoginDetail{UserId: userInfo.Id}).First(&loginDetail).Error; err != nil {
				return nil, err
			}
			if !util.ComparePassword(loginRequest.Password, loginDetail.Password) {
				return nil, ErrInvalidCredential
			}
			return &dto.Authentication{
				Subject:       userInfo.Id,
				Authenticated: true,
				Credential:    loginDetail.Password,
				GrantType:     loginRequest.GrantType,
				ClientId:      loginRequest.ClientId,
			}, nil
		}),
	).EvalWithHandlerE(func(err error) error {
		requestErr := &request.Error{}
		if errors.As(err, &requestErr) || errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvalidCredential
		}
		return nil
	})
}

func (s *Auth) refreshTokenAuthenticated(grantRequest *dto.GrantRequest) (*dto.Authentication, error) {
	refreshToken := &model.RefreshToken{}
	if err := s.db.Where(&model.RefreshToken{Id: grantRequest.RefreshToken}).First(refreshToken).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUnauthorizedAccess
	} else if err != nil {
		return nil, err
	}
	if time.Now().After(refreshToken.ExpiredAt) {
		return nil, ErrUnauthorizedAccess
	}
	return h.FlatMap2(
		h.Lift(base64.StdEncoding.DecodeString)(refreshToken.ProtectedTicket),
		h.Lift(util.UnmarshalJsonStruct[dto.UserInfo]),
		h.Lift(func(userInfo dto.UserInfo) (*dto.Authentication, error) {
			return &dto.Authentication{
				Authenticated: true,
				Subject:       refreshToken.Subject,
				Credential:    refreshToken,
				GrantType:     grantRequest.GrantType,
			}, nil
		}),
	).Eval()
}

func (s *Auth) Logout(refreshToken string) error {
	return s.db.Where(&model.RefreshToken{Id: refreshToken}).Delete(&model.RefreshToken{}).Error
}

func (s *Auth) Activate(userId string) error {
	return s.db.
		Model(&model.LoginDetail{}).
		Where(&model.LoginDetail{UserId: userId}).
		Update("activated", true).Error
}

func (s *Auth) CheckPermission() error {
	return nil
}
