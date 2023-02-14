package controller

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hoquangnam45/pharmacy-auth/dto"
	"github.com/hoquangnam45/pharmacy-auth/service"
	h "github.com/hoquangnam45/pharmacy-common-go/helper/errorHandler"
	"github.com/hoquangnam45/pharmacy-common-go/util"
	"github.com/hoquangnam45/pharmacy-common-go/util/response"
)

type AuthController struct {
	authService *service.Auth
}

func NewAuthController(authService *service.Auth) *AuthController {
	return &AuthController{authService}
}

func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	h.FlatMap4(
		h.Lift(util.ReadAllThenClose)(r.Body),
		h.Lift(util.UnmarshalJsonStructPtr[dto.GrantRequest]),
		h.Lift(c.authService.Login),
		h.Lift(c.authService.GrantAccess),
		h.PeekE(func(grantAccess *dto.GrantAccess) error {
			resp := response.NewJsonResponse(200, grantAccess)
			return response.Handler(resp, w)
		}),
	).EvalWithHandler(func(err error) {
		if err := response.ErrorHandler(err, r, w); err != nil {
			util.Logger.Error(err.Error())
		}
	})
}

func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	h.FlatMap4(
		h.Lift(util.ReadAllThenClose)(r.Body),
		h.Lift(util.UnmarshalJsonStructPtr[dto.GrantRequest]),
		h.Lift(c.authService.Register),
		h.Lift(c.authService.GrantAccess),
		h.PeekE(func(grantAccess *dto.GrantAccess) error {
			return response.Handler(response.NewJsonResponse(200, grantAccess), w)
		}),
	).EvalWithHandler(func(err error) {
		if err := response.ErrorHandler(err, r, w); err != nil {
			util.Logger.Error(err.Error())
		}
	})
}

func (c *AuthController) Activate(w http.ResponseWriter, r *http.Request) {
	h.FactoryM(func() (any, error) {
		userId := chi.URLParam(r, "userId")
		return nil, c.authService.Activate(userId)
	}).EvalWithHandler(func(err error) {
		if err := response.ErrorHandler(err, r, w); err != nil {
			util.Logger.Error(err.Error())
		}
	})
}

func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	h.FlatMap2(
		h.Lift(util.ReadAllThenClose)(r.Body),
		h.Lift(util.UnmarshalJsonStructPtr[dto.RevokeRequest]),
		h.PeekE(func(revokeRequest *dto.RevokeRequest) error {
			return c.authService.Logout(revokeRequest.RefreshToken)
		}),
	).EvalWithHandler(func(err error) {
		if err := response.ErrorHandler(err, r, w); err != nil {
			util.Logger.Error(err.Error())
		}
	})
}

func (c *AuthController) CheckPermission(w http.ResponseWriter, r *http.Request) {
	h.FlatMap2(
		h.Lift(util.ReadAllThenClose)(r.Body),
		h.Lift(util.UnmarshalJsonStructPtr[dto.RevokeRequest]),
		h.PeekE(func(revokeRequest *dto.RevokeRequest) error {
			return c.authService.CheckPermission()
		}),
	).EvalWithHandler(func(err error) {
		if err := response.ErrorHandler(err, r, w); err != nil {
			util.Logger.Error(err.Error())
		}
	})
}
