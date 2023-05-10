package biz

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/hoquangnam45/pharmacy-auth/internal/constant/grantType"
	"github.com/hoquangnam45/pharmacy-auth/internal/constant/oauthProviderType"
	"github.com/hoquangnam45/pharmacy-auth/internal/data"
	"github.com/hoquangnam45/pharmacy-auth/internal/dto"
	"github.com/hoquangnam45/pharmacy-auth/internal/model"
	"github.com/hoquangnam45/pharmacy-common-go/helper/db"
	"github.com/hoquangnam45/pharmacy-common-go/util"
	h "github.com/hoquangnam45/pharmacy-common-go/util/errorHandler"
	"github.com/hoquangnam45/pharmacy-common-go/util/log"
	"github.com/hoquangnam45/pharmacy-common-go/util/request"
	"gorm.io/gorm"
)

type TrustedTpInfoFetcher func(accessToken string) (*dto.UserInfo, error)

type LoginDetailUsecase struct {
	repo             data.LoginDetailRepo
	refreshTokenRepo data.RefreshTokenRepo
	clientRepo       data.ClientRepo
	userInfoClient   UserInfoClient
	data.TransactionManager
	log log.Logger
}

func NewLoginDetailUseCase(userInfoClient UserInfoClient, repo data.LoginDetailRepo, clientRepo data.ClientRepo, refreshTokenRepo data.RefreshTokenRepo, logger log.Logger, transactionManager data.TransactionManager) *LoginDetailUsecase {
	return &LoginDetailUsecase{
		repo:               repo,
		refreshTokenRepo:   refreshTokenRepo,
		clientRepo:         clientRepo,
		log:                logger,
		TransactionManager: transactionManager,
		userInfoClient:     userInfoClient,
	}
}

func (s *LoginDetailUsecase) GenerateAccessToken(refreshToken *model.RefreshToken) (string, error) {
	claims := dto.AuthClaims{}
	client := refreshToken.Client
	claims.ExpiresAt = jwt.NewNumericDate(refreshToken.IssuedAt.Add(client.AccessTokenTtl.ToDuration()))
	claims.IssuedAt = jwt.NewNumericDate(refreshToken.IssuedAt)
	claims.ID = uuid.New().String()
	claims.Issuer = client.Issuer
	claims.Subject = refreshToken.Subject
	signingKey := client.SigningKey
	signingMethod := jwt.GetSigningMethod(client.SigningMethod)
	token := jwt.NewWithClaims(signingMethod, claims)
	privateKey, err := jwt.ParseECPrivateKeyFromPEM([]byte(signingKey))
	if err == nil {
		return token.SignedString(privateKey)
	} else {
		return "", err
	}
}

func (s *LoginDetailUsecase) Activate(id string) error {
	return h.FlatMap2(
		h.Lift(uuid.Parse)(id),
		h.Lift(s.repo.FindByID),
		h.LiftE(func(loginDetail *model.LoginDetail) error {
			if !loginDetail.Activated {
				loginDetail.Activated = true
				return h.Lift(s.repo.Save)(loginDetail).Error()
			}
			return nil
		}),
	).Error()
}

func (s *LoginDetailUsecase) Logout(refreshToken string) error {
	return s.refreshTokenRepo.DeleteById(refreshToken)
}

func (s *LoginDetailUsecase) FindClient(clientId string) (*model.Client, error) {
	return h.Lift(s.clientRepo.FindByClientID)(clientId).EvalWithHandlerE(func(err error) error {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvalidClientId
		}
		return nil
	})
}

func (s *LoginDetailUsecase) Login(grantRequest *dto.GrantRequest) (*dto.Authentication, error) {
	switch grantRequest.GrantType.Normalize() {
	case grantType.Password:
		return s.passwordAuthenticated(grantRequest)
	case grantType.RefreshToken:
		return s.refreshTokenAuthenticated(grantRequest)
	case grantType.TrustedTp:
		switch grantRequest.Provider.Normalize() {
		case oauthProviderType.Facebook:
			return s.trustedTpAuthenticated(FetchFbUserInfo, *grantRequest.AccessToken, grantRequest.ClientId)
		case oauthProviderType.Google:
			return s.trustedTpAuthenticated(FetchGoogleUserInfo, *grantRequest.AccessToken, grantRequest.ClientId)
		default:
			return nil, fmt.Errorf("%w %s", ErrNotSupportOauthProvider, *grantRequest.Provider)
		}
	default:
		return nil, fmt.Errorf("%w %s", ErrNotSupportGrantType, grantRequest.GrantType)
	}
}

func (s *LoginDetailUsecase) Register(registerRequest *dto.GrantRequest) (*dto.Authentication, error) {
	return h.FlatMap2(
		h.FactoryM(func() (*dto.UserInfo, error) {
			return s.userInfoClient.CreateUserInfo(registerRequest.Username, registerRequest.Email, registerRequest.PhoneNumber)
		}),
		h.Lift(func(userInfo *dto.UserInfo) (*model.LoginDetail, error) {
			return h.FlatMap(
				h.Lift(util.HashPassword)(*registerRequest.Password),
				h.Lift(func(pass string) (*model.LoginDetail, error) {
					return s.repo.Save(&model.LoginDetail{
						UserId:    userInfo.Id,
						Password:  pass,
						Activated: false,
					})
				})).Eval()
		}),
		h.LiftJ(func(loginDetail *model.LoginDetail) *dto.Authentication {
			return &dto.Authentication{
				Subject:       loginDetail.UserId,
				Authenticated: true,
				Credential:    loginDetail.Password,
				GrantType:     registerRequest.GrantType,
				ClientId:      registerRequest.ClientId,
			}
		}),
	).EvalWithHandlerE(func(err error) error {
		requestErr := &request.Error{}
		groupErr := &util.GroupError{}
		if !errors.As(err, groupErr) || !errors.Is(groupErr.Group, ErrUserInfoClientGroup) || !errors.Is(groupErr.Cause, ErrResourceAlreadyExists) {
			s.userInfoClient.RemoveUserInfo(registerRequest.Username, registerRequest.Email, registerRequest.Password)
		} else {
			// Credential should already exist here
			return ErrCredentialAlreadyExist
		}
		if db.IsDuplicatedError(err) || errors.As(err, &requestErr) && requestErr.StatusCode == 409 {
			// This should not be happening, it should return err from user info client first, please check the db for inconsistency again
			s.log.Error("inconsistency between auth service and user info service of user[email=%s, username=%s]", registerRequest.Email, registerRequest.Username)
			return ErrCredentialAlreadyExist
		}
		return nil
	})
}

func (s *LoginDetailUsecase) GrantAccess(authentication *dto.Authentication) (*dto.GrantAccess, error) {
	if authentication == nil || !authentication.Authenticated {
		return nil, ErrUnauthorizedAccess
	}

	client, err := h.Lift(s.clientRepo.FindByClientID)(authentication.ClientId).EvalWithHandlerE(func(err error) error {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvalidClientId
		}
		return nil
	})
	if err != nil {
		return nil, err
	} else if !client.Active {
		return nil, ErrUnauthorizedAccess
	}
	switch authentication.GrantType.Normalize() {
	case grantType.Password:
		fallthrough
	case grantType.TrustedTp:
		return s.grantAccessCommon(authentication, client)
	case grantType.RefreshToken:
		return s.grantAccessByRefreshToken(authentication, client)
	default:
		return nil, ErrUnauthorizedAccess
	}
}

func (s *LoginDetailUsecase) grantAccessByRefreshToken(authentication *dto.Authentication, client *model.Client) (*dto.GrantAccess, error) {
	var grantAccess *dto.GrantAccess = nil
	err := s.Run(func() error {
		refreshToken := authentication.Credential.(*model.RefreshToken)
		now := time.Now()
		if err := s.refreshTokenRepo.DeleteById(refreshToken.Id); err != nil {
			return err
		}
		newRefreshToken, err := s.refreshTokenRepo.Save(&model.RefreshToken{
			Id:        base64.StdEncoding.EncodeToString([]byte(uuid.New().String())),
			IssuedAt:  now,
			ExpiredAt: now.Add(client.RefreshTokenTtl.ToDuration()),
			ClientId:  client.Id,
			Client:    client,
			Subject:   authentication.Subject,
		})
		if err != nil {
			return err
		}
		accessToken, err := s.GenerateAccessToken(newRefreshToken)
		if err != nil {
			return err
		}
		grantAccess = &dto.GrantAccess{
			RefreshToken: newRefreshToken.Id,
			AccessToken:  accessToken,
			Subject:      newRefreshToken.Subject,
			IssuedAt:     newRefreshToken.IssuedAt,
			ExpiredAt:    newRefreshToken.ExpiredAt,
			ExpiredIn:    client.AccessTokenTtl.ToDuration(),
			ClientId:     client.ClientId,
		}
		return nil
	})
	return grantAccess, err
}

func (s *LoginDetailUsecase) grantAccessCommon(authentication *dto.Authentication, client *model.Client) (*dto.GrantAccess, error) {
	now := time.Now()
	newRefreshToken := &model.RefreshToken{
		Id:        base64.StdEncoding.EncodeToString([]byte(uuid.New().String())),
		IssuedAt:  now,
		ExpiredAt: now.Add(client.RefreshTokenTtl.ToDuration()),
		Client:    client,
		ClientId:  client.Id,
		Subject:   authentication.Subject,
	}
	accessToken, err := s.GenerateAccessToken(newRefreshToken)
	if err != nil {
		return nil, err
	}
	grantAccess := &dto.GrantAccess{
		RefreshToken: newRefreshToken.Id,
		AccessToken:  accessToken,
		Subject:      newRefreshToken.Subject,
		IssuedAt:     newRefreshToken.IssuedAt,
		ExpiredAt:    newRefreshToken.ExpiredAt,
		ExpiredIn:    client.AccessTokenTtl.ToDuration(),
		ClientId:     client.ClientId,
	}
	if _, err = s.refreshTokenRepo.Save(newRefreshToken); err != nil {
		return nil, err
	}
	return grantAccess, nil
}

func (s *LoginDetailUsecase) trustedTpAuthenticated(fetcher TrustedTpInfoFetcher, accessToken string, clientId string) (*dto.Authentication, error) {
	return h.FlatMap2(
		h.Lift(fetcher)(accessToken),
		h.Lift(func(userInfo *dto.UserInfo) (*dto.UserInfo, error) {
			if userInfo, err := s.userInfoClient.FetchUserInfo(nil, &userInfo.Email, nil); errors.Is(err, ErrResourceAlreadyExists) {
				return s.userInfoClient.CreateUserInfo(nil, &userInfo.Email, nil)
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

func (s *LoginDetailUsecase) passwordAuthenticated(loginRequest *dto.GrantRequest) (*dto.Authentication, error) {
	return h.FlatMap(
		h.FactoryM(func() (*dto.UserInfo, error) {
			return s.userInfoClient.FetchUserInfo(loginRequest.Username, loginRequest.Email, loginRequest.PhoneNumber)
		}),
		h.Lift(func(userInfo *dto.UserInfo) (*dto.Authentication, error) {
			loginDetail, err := s.repo.FindByUserId(userInfo.Id)
			if err != nil {
				return nil, err
			}
			if !util.ComparePassword(*loginRequest.Password, loginDetail.Password) {
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

func (s *LoginDetailUsecase) refreshTokenAuthenticated(grantRequest *dto.GrantRequest) (*dto.Authentication, error) {
	var auth *dto.Authentication = nil
	err := s.Run(func() error {
		refreshToken, err := s.refreshTokenRepo.FindById(*grantRequest.RefreshToken)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUnauthorizedAccess
		} else if err != nil {
			return err
		}
		if time.Now().After(refreshToken.ExpiredAt) {
			s.refreshTokenRepo.DeleteById(refreshToken.Id)
			return ErrUnauthorizedAccess
		}
		authI, err := h.FlatMap2(
			h.Lift(base64.StdEncoding.DecodeString)(refreshToken.ProtectedTicket),
			h.Lift(util.UnmarshalJson(&dto.UserInfo{})),
			h.Lift(func(userInfo *dto.UserInfo) (*dto.Authentication, error) {
				return &dto.Authentication{
					Authenticated: true,
					Subject:       refreshToken.Subject,
					Credential:    refreshToken,
					GrantType:     grantRequest.GrantType,
					ClientId:      grantRequest.ClientId,
				}, nil
			}),
		).Eval()
		if err != nil {
			return err
		}
		auth = authI
		return nil
	})
	return auth, err
}

// TODO: Implement this method
func (s *LoginDetailUsecase) CheckPermission() error {
	return nil
}
