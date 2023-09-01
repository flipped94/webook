package tencent

import (
	"context"
	"fmt"

	tencentsms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"

	"github.com/flipped94/webook/internal/service/sms"
)

type Service struct {
	appId    *string
	signName *string
	client   *tencentsms.Client
}

func NewService(client *tencentsms.Client, appId string, signName string) *Service {
	return &Service{
		client:   client,
		appId:    &appId,
		signName: &signName,
	}
}

func (s *Service) Send(ctx context.Context, template string, args []sms.NamedArg, numbers []string) error {
	req := tencentsms.NewSendSmsRequest()
	req.SmsSdkAppId = s.appId
	req.SignName = s.signName
	req.TemplateId = &template
	req.TemplateParamSet = s.toStringArgPtrSlice(args)
	req.PhoneNumberSet = s.toStringPtrSlice(numbers)
	resp, err := s.client.SendSms(req)
	if err != nil {
		return err
	}
	for _, status := range resp.Response.SendStatusSet {
		if status.Code == nil || *(status.Code) != "Ok" {
			return fmt.Errorf("发送短信失败 %s, %s ", *status.Code, *status.Message)
		}
	}
	return nil
}

func (s *Service) toStringPtrSlice(source []string) []*string {
	res := make([]*string, 0, len(source))
	for i := 0; i < len(source); i++ {
		res = append(res, &source[i])
	}
	return res
}

func (s *Service) toStringArgPtrSlice(source []sms.NamedArg) []*string {
	res := make([]*string, 0, len(source))
	for i := 0; i < len(source); i++ {
		res = append(res, &source[i].Value)
	}
	return res
}
