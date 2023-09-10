package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/flipped94/webook/internal/service/sms"
	smsmocks "github.com/flipped94/webook/internal/service/sms/mocks"
	"github.com/flipped94/webook/pkg/ratelimit"
	limitmocks "github.com/flipped94/webook/pkg/ratelimit/mocks"
)

func TestRateLimitSmsService_Send(t *testing.T) {
	testCases := []struct {
		name      string
		mock      func(ctrl *gomock.Controller) (sms.Service, ratelimit.Limiter)
		tpl       string
		args      []sms.NamedArg
		numbers   []string
		wantError error
	}{
		{
			name: "限流成功",
			mock: func(ctrl *gomock.Controller) (sms.Service, ratelimit.Limiter) {
				limiter := limitmocks.NewMockLimiter(ctrl)
				service := smsmocks.NewMockService(ctrl)
				limiter.EXPECT().Limited(gomock.Any(), "sms").Return(true, nil)
				return service, limiter
			},
			tpl: "TestSMS",
			args: []sms.NamedArg{
				{Name: "code", Value: "1234"},
			},
			numbers:   []string{"13612345678"},
			wantError: fmt.Errorf("触发限流"),
		},
		{
			name: "限流异常",
			mock: func(ctrl *gomock.Controller) (sms.Service, ratelimit.Limiter) {
				limiter := limitmocks.NewMockLimiter(ctrl)
				service := smsmocks.NewMockService(ctrl)
				limiter.EXPECT().Limited(gomock.Any(), "sms").Return(false, errors.New("限流异常"))
				return service, limiter
			},
			tpl: "TestSMS",
			args: []sms.NamedArg{
				{Name: "code", Value: "1234"},
			},
			numbers:   []string{"13612345678"},
			wantError: fmt.Errorf("短信服务判断限流出现问题"),
		},
		{
			name: "短信服务异常",
			mock: func(ctrl *gomock.Controller) (sms.Service, ratelimit.Limiter) {
				limiter := limitmocks.NewMockLimiter(ctrl)
				service := smsmocks.NewMockService(ctrl)
				service.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("短信服务异常"))
				limiter.EXPECT().Limited(gomock.Any(), gomock.Any()).Return(false, nil)
				return service, limiter
			},
			tpl: "TestSMS",
			args: []sms.NamedArg{
				{Name: "code", Value: "1234"},
			},
			numbers:   []string{"13612345678"},
			wantError: fmt.Errorf("短信服务异常"),
		},
		{
			name: "发送成功",
			mock: func(ctrl *gomock.Controller) (sms.Service, ratelimit.Limiter) {
				limiter := limitmocks.NewMockLimiter(ctrl)
				service := smsmocks.NewMockService(ctrl)
				service.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				limiter.EXPECT().Limited(gomock.Any(), gomock.Any()).Return(false, nil)
				return service, limiter
			},
			tpl: "TestSMS",
			args: []sms.NamedArg{
				{Name: "code", Value: "1234"},
			},
			numbers:   []string{"13612345678"},
			wantError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()
			service := NewRateLimitSmsService(tc.mock(controller))
			err := service.Send(context.Background(), tc.tpl, tc.args, tc.numbers)
			assert.Equal(t, tc.wantError, err)
		})
	}
}
