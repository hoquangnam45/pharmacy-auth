package oauthProviderType

import "strings"

type OAuthProvider string

const (
	Invalid  OAuthProvider = ""
	Facebook OAuthProvider = "FACEBOOK"
	Google   OAuthProvider = "GOOGLE"
)

func (s OAuthProvider) Normalize() OAuthProvider {
	providerType := OAuthProvider(strings.ToUpper(strings.Trim(string(s), " ")))
	if !providerType.IsValid() {
		return Invalid
	}
	return providerType
}

func (s OAuthProvider) IsValid() bool {
	switch s {
	case Facebook, Google:
		return true
	}
	return false
}

func (s OAuthProvider) Value() string {
	return string(s)
}

func FromString(s string) OAuthProvider {
	return OAuthProvider(s).Normalize()
}

func FromStringPtr(s *string) *OAuthProvider {
	if s == nil {
		return nil
	}
	ret := FromString(*s)
	return &ret
}
