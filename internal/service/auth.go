package service

import (
	"context"

	v1 "github.com/hoquangnam45/pharmacy-auth/api/auth/v1"
	"github.com/hoquangnam45/pharmacy-auth/internal/biz"
	"github.com/hoquangnam45/pharmacy-auth/internal/constant/grantType"
	"github.com/hoquangnam45/pharmacy-auth/internal/constant/oauthProviderType"
	h "github.com/hoquangnam45/pharmacy-common-go/util/errorHandler"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Auth struct {
	uc *biz.LoginDetailUsecase
	v1.UnimplementedAuthServer
}

func NewAuthService(uc *biz.LoginDetailUsecase) *Auth {
	return &Auth{uc: uc}
}

func (s *Auth) Token(ctx context.Context, grantRequest *v1.GrantRequest) (*v1.GrantAccess, error) {
	return h.FlatMap3(
		h.LiftJ(mapApiGrantRequest)(grantRequest),
		h.Lift(s.uc.Login),
		h.Lift(s.uc.GrantAccess),
		h.LiftJ(mapBizGrantAccess),
	).EvalWithContext(ctx)
}

func (s *Auth) Register(ctx context.Context, registerRequest *v1.GrantRequest) (*v1.GrantAccess, error) {
	return h.FlatMap3(
		h.LiftJ(mapApiGrantRequest)(registerRequest),
		h.Lift(s.uc.Register),
		h.Lift(s.uc.GrantAccess),
		h.LiftJ(mapBizGrantAccess),
	).EvalWithContext(ctx)
}

func (s *Auth) Logout(ctx context.Context, logoutRequest *v1.LogoutRequest) (*emptypb.Empty, error) {
	return h.FlatMap(
		h.LiftE(s.uc.Logout)(logoutRequest.RefreshToken),
		h.LiftJ(empty[any]),
	).EvalWithContext(ctx)
}

func (s *Auth) Activate(ctx context.Context, activateRequest *v1.ActivateRequest) (*emptypb.Empty, error) {
	return h.FlatMap(
		h.LiftE(s.uc.Activate)(activateRequest.UserId),
		h.LiftJ(empty[any]),
	).EvalWithContext(ctx)
}

func (s *Auth) CheckPermission() error {
	return s.uc.CheckPermission()
}

func mapBizGrantAccess(req *biz.GrantAccess) *v1.GrantAccess {
	return &v1.GrantAccess{
		AccessToken:  req.AccessToken,
		RefreshToken: req.RefreshToken,
		Subject:      req.Subject,
		ExpiredIn:    durationpb.New(req.ExpiredIn),
		IssuedAt:     timestamppb.New(req.IssuedAt),
		ExpiredAt:    timestamppb.New(req.ExpiredAt),
		ClientId:     req.ClientId,
	}
}

func mapApiGrantRequest(req *v1.GrantRequest) *biz.GrantRequest {
	return &biz.GrantRequest{
		GrantType: grantType.GrantType(req.GrantType).Normalize(),
		ClientId:  req.ClientId,
		TrustedThirdPartyGrantRequest: &biz.TrustedThirdPartyGrantRequest{
			Provider:    oauthProviderType.FromStringPtr(req.Provider),
			AccessToken: req.AccessToken,
		},
		PasswordGrantRequest: &biz.PasswordGrantRequest{
			Username:    req.Username,
			Password:    req.Password,
			PhoneNumber: req.PhoneNumber,
			Email:       req.Email,
		},
		RefreshGrantRequest: &biz.RefreshGrantRequest{
			RefreshToken: req.RefreshToken,
		},
	}
}

func empty[T any](in T) *emptypb.Empty {
	return &emptypb.Empty{}
}

type Valueable interface {
	Value() string
}

func StringPtr[T Valueable](a *T) *string {
	if a == nil {
		return nil
	}
	r := *a
	str := r.Value()
	return &str
}
