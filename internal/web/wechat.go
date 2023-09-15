package web

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/flipped94/webook/internal/service"
	"github.com/flipped94/webook/internal/service/oauth2/wechat"
	jwt2 "github.com/flipped94/webook/internal/web/jwt"
)

type OAuth2WechatHandler struct {
	svc     wechat.Service
	cmd     redis.Cmdable
	userSvc service.UserService
	jwt2.Handler
}

func NewOauth2WechatHandler(svc wechat.Service, userSvc service.UserService, jwtHandler jwt2.Handler, cmd redis.Cmdable) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:     svc,
		cmd:     cmd,
		userSvc: userSvc,
		Handler: jwtHandler,
	}
}

func (h *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	group := server.Group("/oauth/wechat")
	group.GET("/qrcode", h.QrCode)
	group.Any("/callback", h.Callback)
	group.GET("/heartbeat", h.HeartBeat)
}

func (h *OAuth2WechatHandler) QrCode(ctx *gin.Context) {
	sceneStr := uuid.New().String()
	response, err := h.svc.QrCodeStream(ctx, sceneStr)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	stream, err := io.ReadAll(response.Body)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.Header("X-WX-Heart-Beat", sceneStr)
	_, err = ctx.Writer.Write(stream)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
}

func (h *OAuth2WechatHandler) Callback(ctx *gin.Context) {
	echostr := ctx.Query("echostr")
	if echostr != "" {
		ctx.String(http.StatusOK, echostr)
	} else {
		type WechatEvent struct {
			ToUserName   string `json:"to_user_name"`
			FromUserName string `json:"from_user_name"`
			CreateTime   int64  `json:"create_time"`
			MsgType      string `json:"msg_type"`
			Content      string `json:"content"`
			MsgId        int64  `json:"msg_id"`
			Event        string `json:"event"`
			EventKey     string `json:"event_key"` // 扫码事件
		}
		var event WechatEvent
		err := ctx.ShouldBindXML(&event)
		if err != nil {
			fmt.Printf("微信服务器回调参数绑定错误: %s", err.Error())
			return
		}
		// 扫码关注事件 包含为扫码后关注事件和已关注扫码事件
		var param string
		if event.Event == "subscribe" {
			eventKey := strings.Split(event.EventKey, "qrscene_")
			param = eventKey[1]
		} else if event.Event == "SCAN" {
			param = event.EventKey
		}
		openId := event.FromUserName
		userInfo, err := h.svc.WxUserInfo(ctx, openId)
		if err != nil {
			fmt.Printf("没有获取到微信信息: %s", err.Error())
			return
		}
		u, err := h.userSvc.FindOrCreateByWechat(ctx, userInfo)
		if err != nil {
			fmt.Printf("openid创建用户错误: %s", err.Error())
			return
		}

		err = h.cmd.Set(ctx, param+":userId", u.Id, time.Minute*10).Err()
		if err != nil {
			fmt.Printf("微信扫码创建userId缓存失败: %s", err.Error())
			return
		}
	}
}

func (h *OAuth2WechatHandler) HeartBeat(ctx *gin.Context) {
	key := ctx.Query("key")
	userId, err := h.cmd.Get(ctx, key+":userId").Int64()
	if err != nil {
		fmt.Printf("查询微信扫码创建userId缓存失败: %s", err.Error())
		return
	}
	u, err := h.userSvc.Profile(ctx, userId)
	if err != nil {
		fmt.Printf("扫码心跳没有查询到用户: %s", err.Error())
		return
	}
	err = h.SetLoginToken(ctx, userId)
	if err != nil {
		fmt.Printf("心跳设置LoginToken失败: %s", err.Error())
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: http.StatusOK,
		Msg:  "success",
		Data: u,
	})
}
