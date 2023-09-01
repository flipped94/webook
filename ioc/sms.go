package ioc

import (
	"github.com/flipped94/webook/internal/service/sms"
	"github.com/flipped94/webook/internal/service/sms/memory"
)

func InitSmsService() sms.Service {
	return memory.NewService()
}
