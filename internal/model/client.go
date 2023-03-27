package model

import "time"

type Client struct {
	Id              *int `gorm:"primaryKey"`
	ClientId        string
	SigningKey      string
	VerificationKey string
	SigningMethod   string
	Issuer          string
	Active          bool
	RefreshTokenTtl Duration
	AccessTokenTtl  Duration
	AllowOrigin     string
}

type Duration int64

func (d *Duration) ToDuration() time.Duration {
	return time.Second * time.Duration(int64(*d))
}
