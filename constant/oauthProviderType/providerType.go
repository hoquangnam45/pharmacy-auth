package oauthProviderType

import "strings"

type OAuthProvider string

const (
	Facebook OAuthProvider = "FACEBOOK"
	Google   OAuthProvider = "GOOGLE"
)

func (s OAuthProvider) Normalize() OAuthProvider {
	return OAuthProvider(strings.ToUpper(strings.Trim(string(s), " ")))
}

func (s OAuthProvider) IsValid() bool {
	switch s.Normalize() {
	case Facebook, Google:
		return true
	}
	return false
}
