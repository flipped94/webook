package repository

import (
	"context"

	"github.com/flipped94/webook/internal/repository/cache"
)

var (
	ErrCodeSendTooMany        = cache.ErrCodeSendTooMany
	ErrCodeVerifyTooManyTimes = cache.ErrCodeVerifyTooManyTimes
)

type CodeRepository interface {
	Store(ctx context.Context, biz string, phone string, code string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
}

type CachedCodeRepository struct {
	cache cache.CodeCache
}

func NewCodeRepository(cache cache.CodeCache) CodeRepository {
	return &CachedCodeRepository{
		cache: cache,
	}
}

func (repo *CachedCodeRepository) Store(ctx context.Context, biz string, phone string, code string) error {
	return repo.cache.Set(ctx, biz, phone, code)
}

func (repo *CachedCodeRepository) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	return repo.cache.Verify(ctx, biz, phone, inputCode)
}
