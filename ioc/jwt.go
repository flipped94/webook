package ioc

import (
	"github.com/redis/go-redis/v9"

	jwt2 "github.com/flipped94/webook/internal/web/jwt"
)

func InitJwtHandler(cmd redis.Cmdable) jwt2.Handler {
	return jwt2.NewRedisJWTHandler(cmd)
}
