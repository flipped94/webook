package ratelimit

import (
	"context"
	"fmt"
	"github.com/flipped94/webook/internal/service/sms"
	"github.com/flipped94/webook/pkg/ratelimit"
)

type RateLimitSmsService struct {
	svc     sms.Service
	limiter ratelimit.Limiter
}

func NewRateLimitSmsService(svc sms.Service, limiter ratelimit.Limiter) sms.Service {
	return &RateLimitSmsService{
		svc:     svc,
		limiter: limiter,
	}
}

func (s *RateLimitSmsService) Send(ctx context.Context, template string, args []sms.NamedArg, numbers []string) error {
	// 限流
	limited, err := s.limiter.Limited(ctx, "sms")
	if err != nil {
		return fmt.Errorf("短信服务判断限流出现问题")
	}
	if limited {
		return fmt.Errorf("触发限流")
	}
	err = s.svc.Send(ctx, template, args, numbers)
	return err
}
