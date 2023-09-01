package redis

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/flipped94/webook/internal/repository/cache"
)

//go:embed lua/set_code.lua
var luaSetCode string

//go:embed lua/verify_code.lua
var luaVerifyCode string

type RedisCodeCache struct {
	client redis.Cmdable
}

func NewCodeCache(client redis.Cmdable) cache.CodeCache {
	return &RedisCodeCache{
		client: client,
	}
}

func (c *RedisCodeCache) Set(ctx context.Context, biz string, phone string, code string) error {
	res, err := c.client.Eval(ctx, luaSetCode, []string{c.key(biz, phone)}, code).Int()
	if err != nil {
		return err
	}
	switch res {
	case 0:
		return nil
	case -1:
		// 发送太频繁
		return cache.ErrCodeSendTooMany
	default:
		// 系统错误
		return errors.New("系统错误")
	}
}

func (c *RedisCodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	res, err := c.client.Eval(ctx, luaVerifyCode, []string{c.key(biz, phone)}, inputCode).Int()
	if err != nil {
		return false, err
	}
	switch res {
	case 0:
		return true, nil
	case -1:
		// 正常来说，如果频繁出现这个错误，你就要告警，因为有人搞你
		return false, cache.ErrCodeVerifyTooManyTimes
	case -2:
		return false, nil
	}
	return false, cache.ErrUnknownForCode
}

func (c *RedisCodeCache) key(biz string, phone string) string {
	s := fmt.Sprintf("phone_code:%s:%s", biz, phone)
	return s
}
