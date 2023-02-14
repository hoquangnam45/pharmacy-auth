package model

import (
	"time"
)

type Client struct {
	Id              int `gorm:"primaryKey"`
	ClientId        string
	SigningKey      string
	Method          string
	VerificationKey string
	Name            *string
	ApplicationType int
	Active          bool
	RefreshTokenTtl time.Duration
	AccessTokenTtl  time.Duration
	AllowOrigin     string
	Issuer          string
}
