package model

import (
	"github.com/google/uuid"
)

type Config struct {
	Id        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()"`
	Namespace *string
	Key       string
	Value     *string
}
