package model

import (
	"time"
)

type RefreshToken struct {
	Id              string  `gorm:"type:uuid;default:gen_random_uuid()"`
	Client          *Client `gorm:"foreignKey:client_id"`
	Subject         string
	IssuedAt        time.Time
	ExpiredAt       time.Time
	ProtectedTicket string
}
