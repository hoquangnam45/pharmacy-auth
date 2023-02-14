package grantType

import "strings"

type GrantType string

const (
	RefreshToken GrantType = "REFRESH_TOKEN"
	Password     GrantType = "PASSWORD"
	TrustedTp    GrantType = "TRUSTED_TP"
)

func (s GrantType) Normalize() GrantType {
	return GrantType(strings.ToUpper(strings.Trim(string(s), " ")))
}

func (s GrantType) IsValid() bool {
	switch s.Normalize() {
	case RefreshToken, Password, TrustedTp:
		return true
	}
	return false
}
