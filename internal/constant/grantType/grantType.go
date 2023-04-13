package grantType

import "strings"

type GrantType string

const (
	Invalid      GrantType = ""
	RefreshToken GrantType = "REFRESH_TOKEN"
	Password     GrantType = "PASSWORD"
	TrustedTp    GrantType = "TRUSTED_TP"
)

func (s GrantType) Normalize() GrantType {
	grantType := GrantType(strings.ToUpper(strings.Trim(string(s), " ")))
	if !grantType.IsValid() {
		return Invalid
	}
	return grantType
}

func (s GrantType) IsValid() bool {
	switch s {
	case RefreshToken, Password, TrustedTp:
		return true
	}
	return false
}

func (s GrantType) Value() string {
	return string(s)
}

func FromString(s string) GrantType {
	return GrantType(s).Normalize()
}

func FromStringPtr(s *string) *GrantType {
	if s == nil {
		return nil
	}
	ret := FromString(*s)
	return &ret
}
