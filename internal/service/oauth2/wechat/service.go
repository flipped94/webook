package wechat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type WxUserInfo struct {
	Openid  string `json:"openid"`
	Unionid string `json:"unionid"`
}

type Service interface {
	QrCodeStream(ctx context.Context, sceneStr string) (*http.Response, error)
	WxUserInfo(ctx context.Context, openid string) (WxUserInfo, error)
}

type service struct {
	appid     string
	appsecret string
	client    *http.Client
	cmd       redis.Cmdable
}

func NewService(appid string, appsecret string, client *http.Client, cmd redis.Cmdable) Service {
	return &service{
		appid:     appid,
		appsecret: appsecret,
		client:    client,
		cmd:       cmd,
	}
}

func (s *service) QrCodeStream(ctx context.Context, sceneStr string) (*http.Response, error) {
	token, err := s.accessToken(ctx)
	if err != nil {
		return nil, err
	}
	ticket, err := s.qrCodeTicket(ctx, token, sceneStr)
	if err != nil {
		return nil, err
	}
	response, err := s.qrCodeStream(ctx, ticket)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (s *service) accessToken(ctx context.Context) (string, error) {
	key := "wechat:oauth:accesstoken"
	accessToken, err := s.cmd.Get(ctx, key).Result()
	if err != nil {
		const accessTokenPattern = "https://api.weixin.qq.com/cgi-bin/stable_token"
		type StableAccessTokenReq struct {
			GrantType string `json:"grant_type"`
			Appid     string `json:"appid"`
			Secret    string `json:"secret"`
		}
		stableTokenReq := StableAccessTokenReq{
			GrantType: "client_credential",
			Appid:     s.appid,
			Secret:    s.appsecret,
		}
		body, err := json.Marshal(stableTokenReq)
		if err != nil {
			return "", errors.New("系统异常")
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, accessTokenPattern, bytes.NewReader(body))
		if err != nil {
			return "", errors.New("系统错误")
		}
		response, err := s.client.Do(req)
		if err != nil {
			return "", errors.New("系统错误")
		}
		decoder := json.NewDecoder(response.Body)
		var res AccessTokenResp
		err = decoder.Decode(&res)
		if err != nil || res.AccessToken == "" {
			return "", errors.New("系统错误")
		}
		s.cmd.Set(ctx, key, res.AccessToken, time.Hour*2)
		return accessToken, nil
	}
	return accessToken, nil
}

func (s *service) qrCodeTicket(ctx context.Context, accessToken string, sceneStr string) (string, error) {
	const qrCodeTicketPattern = "https://api.weixin.qq.com/cgi-bin/qrcode/create?access_token=%s"
	target := fmt.Sprintf(qrCodeTicketPattern, accessToken)
	reqData := QrCodeTicketReq{
		ExpireSeconds: 604800,
		ActionName:    "QR_STR_SCENE",
		ActionInfo: ActionInfo{
			Scene: Scene{
				SceneStr: sceneStr,
			},
		},
	}
	body, err := json.Marshal(reqData)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewBuffer(body))
	response, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	decoder := json.NewDecoder(response.Body)
	var ticketResp TicketResp
	err = decoder.Decode(&ticketResp)
	if err != nil || ticketResp.Ticket == "" {
		return "", errors.New("系统错误")
	}
	return ticketResp.Ticket, nil
}

func (s *service) qrCodeStream(ctx context.Context, ticket string) (*http.Response, error) {
	const qrCodePattern = "https://mp.weixin.qq.com/cgi-bin/showqrcode?ticket=%s"
	target := fmt.Sprintf(qrCodePattern, ticket)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	response, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (s *service) WxUserInfo(ctx context.Context, openid string) (WxUserInfo, error) {
	accessToken, err := s.accessToken(ctx)
	if err != nil {
		return WxUserInfo{}, err
	}
	const userInfoPattern = "https://api.weixin.qq.com/cgi-bin/user/info?access_token=%s&openid=%s&lang=zh_CN"
	target := fmt.Sprintf(userInfoPattern, accessToken, openid)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	response, err := s.client.Do(req)
	if err != nil {
		return WxUserInfo{}, err
	}
	decoder := json.NewDecoder(response.Body)
	var userInfo WxUserInfo
	err = decoder.Decode(&userInfo)
	if err != nil {
		return WxUserInfo{}, err
	}
	return userInfo, nil
}

type AccessTokenResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type QrCodeTicketReq struct {
	ExpireSeconds int        `json:"expire_seconds"`
	ActionName    string     `json:"action_name"`
	ActionInfo    ActionInfo `json:"action_info"`
}

type ActionInfo struct {
	Scene Scene `json:"scene"`
}

type Scene struct {
	SceneStr string `json:"scene_str"`
}

type TicketResp struct {
	Ticket        string `json:"ticket"`
	ExpireSeconds int    `json:"expire_seconds"`
	Url           string `json:"url"`
}
