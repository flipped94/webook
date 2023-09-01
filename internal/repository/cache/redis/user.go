package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/flipped94/webook/internal/domain"
	"github.com/flipped94/webook/internal/repository/cache"
)

type RedisUserCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

func NewUserCache(client redis.Cmdable) cache.UserCache {
	return &RedisUserCache{
		client:     client,
		expiration: time.Minute * 15,
	}
}

func (cache *RedisUserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	sc := cache.client.Get(ctx, cache.key(id))
	bytes, err := sc.Bytes()
	if err != nil {
		return domain.User{}, err
	}
	var u domain.User
	err = json.Unmarshal(bytes, &u)
	return u, err
}

func (cache *RedisUserCache) Set(ctx context.Context, u domain.User) error {
	bytes, err := json.Marshal(u)
	if err != nil {
		return err
	}
	cache.client.Set(ctx, cache.key(u.Id), bytes, cache.expiration)
	return nil
}

func (cache *RedisUserCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}
