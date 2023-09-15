package ioc

import (
	"errors"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"

	"github.com/flipped94/webook/internal/service/oauth2/wechat"
)

func InitOAuth2WechatService(cmd redis.Cmdable) wechat.Service {
	openid, ok := os.LookupEnv("OPENID")
	if !ok {
		panic(errors.New("初始化微信登录失败"))
	}
	appsecret, ok := os.LookupEnv("APPSECRET")
	if !ok {
		panic(errors.New("初始化微信登录失败"))
	}
	return wechat.NewService(openid, appsecret, http.DefaultClient, cmd)
}
