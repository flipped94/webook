package aliyun

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	dysmsapi20170525 "github.com/alibabacloud-go/dysmsapi-20170525/v3/client"

	"github.com/flipped94/webook/internal/service/sms"
)

type Service struct {
	signName string
	client   *dysmsapi20170525.Client
}

func NewService(client *dysmsapi20170525.Client, signName string) *Service {
	return &Service{
		client:   client,
		signName: signName,
	}
}

func (s *Service) Send(ctx context.Context, template string, args []sms.NamedArg, numbers []string) error {
	argPtr, err := s.toArgsPtr(args)
	if err != nil {
		return err
	}
	sendSmsRequest := &dysmsapi20170525.SendSmsRequest{
		SignName:      &s.signName,
		TemplateCode:  &template,
		PhoneNumbers:  s.toStringPtr(numbers),
		TemplateParam: argPtr,
	}
	resp, err := s.client.SendSms(sendSmsRequest)
	if err != nil {
		return err
	}
	if resp.Body.Code == nil || *(resp.Body.Code) != "OK" {
		return fmt.Errorf("发送短信失败 %s, %s ", *(resp.Body.Code), *(resp.Body.Message))
	}
	return nil
}

func (s *Service) toStringPtr(source []string) *string {
	str := strings.Join(source, ",")
	return &str
}

func (s *Service) toArgsPtr(source []sms.NamedArg) (*string, error) {
	argMap := make(map[string]string, len(source))
	for _, arg := range source {
		argMap[arg.Name] = arg.Value
	}
	bytes, err := json.Marshal(argMap)
	if err != nil {
		return nil, err
	}
	str := string(bytes)
	return &str, nil
}
