package aliyun

import (
	"context"
	"os"
	"testing"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi20170525 "github.com/alibabacloud-go/dysmsapi-20170525/v3/client"
	"github.com/go-playground/assert/v2"

	"github.com/flipped94/webook/internal/service/sms"
)

func TestService_SendSms(t *testing.T) {

	accessKeyId, ok := os.LookupEnv("ACCESS_KEY_ID")
	if !ok {
		t.Fatal()
	}
	accessKeySecret, ok := os.LookupEnv("ACCESS_KEY_SECRET")
	if !ok {
		t.Fatal()
	}
	endpoint := "dysmsapi.aliyuncs.com"

	config := &openapi.Config{
		AccessKeyId:     &accessKeyId,
		AccessKeySecret: &accessKeySecret,
		Endpoint:        &endpoint,
	}
	client, err := dysmsapi20170525.NewClient(config)
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(client, "阿里云短信测试")

	tests := []struct {
		name     string
		template string
		phone    []string
		args     []sms.NamedArg
		wantErr  error
	}{
		{
			name:     "阿里云短信测试",
			template: "SMS_154950909",
			args: []sms.NamedArg{
				{
					Name:  "code",
					Value: "1234",
				},
			},
			phone: []string{""},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			er := service.Send(context.Background(), tc.template, tc.args, tc.phone)
			assert.Equal(t, tc.wantErr, er)
		})
	}
}
