package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	gocache "github.com/fanjindong/go-cache"

	"github.com/flipped94/webook/internal/repository/cache"
)

var lock sync.Mutex

type MemoryCodeCache struct {
	client gocache.ICache
}

func NewCodeCache(cache gocache.ICache) cache.CodeCache {
	return &MemoryCodeCache{
		client: cache,
	}
}

func (c *MemoryCodeCache) Set(ctx context.Context, biz string, phone string, code string) error {
	lock.Lock()
	defer lock.Unlock()
	codeKey, cntKey := c.key(biz, phone)
	_, exist := c.client.Get(codeKey)
	remain, b := c.client.Ttl(cntKey)
	if exist && !b {
		return errors.New("系统错误")
	}
	if !exist || remain < time.Minute*9 {
		c.client.Set(codeKey, code, gocache.WithEx(time.Minute*10))
		c.client.Set(cntKey, int64(3), gocache.WithEx(time.Minute*10))
		return nil
	}
	return cache.ErrCodeSendTooMany
}

func (c *MemoryCodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	lock.Lock()
	defer lock.Unlock()
	codeKey, cntKey := c.key(biz, phone)
	codeValue, codeExist := c.client.Get(codeKey)
	cntValue, cntExist := c.client.Get(cntKey)

	if !codeExist {
		return false, nil
	}
	if !cntExist {
		return false, nil
	}
	cnt := cntValue.(int64)
	if cnt <= 0 {
		return false, cache.ErrCodeVerifyTooManyTimes
	}
	code := codeValue.(string)
	if inputCode == code {
		c.client.Set(cntKey, int64(-1))
		return true, nil
	}
	return false, nil
}

func (c *MemoryCodeCache) key(biz string, phone string) (string, string) {
	codeKey := fmt.Sprintf("phone_code:%s:%s", biz, phone)
	cntKey := fmt.Sprintf("%s:%s", codeKey, "cnt")
	return codeKey, cntKey
}
