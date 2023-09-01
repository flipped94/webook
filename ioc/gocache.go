package ioc

import "github.com/fanjindong/go-cache"

func InitGoCache() cache.ICache {
	return cache.NewMemCache()
}
