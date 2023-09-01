package service

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/flipped94/webook/internal/repository"
	"github.com/flipped94/webook/internal/service/sms"
)

var (
	ErrCodeVerifyTooManyTimes = repository.ErrCodeVerifyTooManyTimes
	ErrCodeSendTooMany        = repository.ErrCodeSendTooMany
)

type CodeService interface {
	Send(ctx context.Context, template string, biz string, phone string) error
	Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error)
}

type codeService struct {
	repo   repository.CodeRepository
	smsSvc sms.Service
}

func NewCodeService(repo repository.CodeRepository, smsSvc sms.Service) CodeService {
	return &codeService{
		repo:   repo,
		smsSvc: smsSvc,
	}
}

func (svc *codeService) Send(ctx context.Context, template string, biz string, phone string) error {
	// 生成一个验证码
	code := svc.generateCode()
	// 缓存
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	err = svc.smsSvc.Send(ctx, template, []sms.NamedArg{{Name: "code", Value: code}}, []string{phone})
	return err
}

func (svc *codeService) Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error) {
	return svc.repo.Verify(ctx, biz, phone, inputCode)
}

func (svc *codeService) generateCode() string {
	num := rand.Intn(1000000)
	// 不够六位的，加上前导 0
	return fmt.Sprintf("%06d", num)
}
