package model

import (
	"time"
)

type RefreshToken struct {
	Id              string `gorm:"primaryKey"`
	ClientId        *int
	Client          *Client
	Subject         string
	IssuedAt        time.Time
	ExpiredAt       time.Time
	ProtectedTicket string
}
