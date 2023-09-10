package ratelimit

import "context"

type Limiter interface {
	// Limited 是否限流. key 限流对象
	// bool 代表是否限流，tru e代表限流
	// error 限流器本身是否有错
	Limited(ctx context.Context, key string) (bool, error)
}
