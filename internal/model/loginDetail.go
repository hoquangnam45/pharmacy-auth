package model

import "github.com/google/uuid"

type LoginDetail struct {
	Id        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()"`
	UserId    string
	Password  string
	Activated bool
}
