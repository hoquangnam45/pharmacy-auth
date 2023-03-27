package service

import "github.com/google/wire"

var ProviderSet = wire.NewSet(NewRedisClient, NewMailService, NewConfigService, NewAuthService)
