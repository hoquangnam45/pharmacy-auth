package dto

import (
	"github.com/hoquangnam45/pharmacy-auth/internal/constant/grantType"
)

type Authentication struct {
	Authenticated bool
	Subject       string
	Credential    any
	GrantType     grantType.GrantType
	ClientId      string
}
