package biz

import (
	"time"
)

type GrantAccess struct {
	AccessToken  string        `json:"accessToken"`
	RefreshToken string        `json:"refreshToken"`
	Subject      string        `json:"subject"`
	ExpiredIn    time.Duration `json:"expiredIn"`
	IssuedAt     time.Time     `json:"issuedAt"`
	ExpiredAt    time.Time     `json:"expiredAt"`
	ClientId     string        `json:"clientId"`
}
