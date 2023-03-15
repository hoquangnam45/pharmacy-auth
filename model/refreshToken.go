package model

import (
	"time"
)

type RefreshToken struct {
	Id              string `gorm:"type:uuid;default:gen_random_uuid()"`
	ClientId        string
	Client          Client `gorm:"foreignKey:ClientIdreferences:Id"`
	Subject         string
	IssuedAt        time.Time
	ExpiredAt       time.Time
	ProtectedTicket string
}
