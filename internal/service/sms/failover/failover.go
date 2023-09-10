package failover

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/flipped94/webook/internal/service/sms"
)

type FailoverSmsService struct {
	svcs []sms.Service
	idx  uint64
}

func (f *FailoverSmsService) Send(ctx context.Context, template string, args []sms.NamedArg, numbers []string) error {
	idx := atomic.AddUint64(&f.idx, 1)
	length := uint64(len(f.svcs))
	for i := idx; i < idx+length; i++ {
		svc := f.svcs[i%length]
		err := svc.Send(ctx, template, args, numbers)
		switch err {
		case nil:
			return nil
		case context.DeadlineExceeded:
			return err
		}
	}
	return errors.New("全部短信服务商不可用")
}
