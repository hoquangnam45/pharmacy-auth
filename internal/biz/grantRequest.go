package biz

import (
	"github.com/hoquangnam45/pharmacy-auth/internal/constant/grantType"
	"github.com/hoquangnam45/pharmacy-auth/internal/constant/oauthProviderType"
)

type GrantRequest struct {
	*TrustedThirdPartyGrantRequest
	*PasswordGrantRequest
	*RefreshGrantRequest
	GrantType grantType.GrantType `json:"grantType"`
	ClientId  string              `json:"clientId"`
}

type TrustedThirdPartyGrantRequest struct {
	Provider    *oauthProviderType.OAuthProvider `json:"provider"`
	AccessToken *string                          `json:"accessToken"`
}

type PasswordGrantRequest struct {
	Username    *string `json:"username"`
	Password    *string `json:"password"`
	PhoneNumber *string `json:"phoneNumber"`
	Email       *string `json:"email"`
}

type RefreshGrantRequest struct {
	RefreshToken *string
}
