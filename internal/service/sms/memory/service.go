package memory

import (
	"context"
	"fmt"

	"github.com/flipped94/webook/internal/service/sms"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Send(ctx context.Context, template string, args []sms.NamedArg, numbers []string) error {
	fmt.Println(args)
	return nil
}
