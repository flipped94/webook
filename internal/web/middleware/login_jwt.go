package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/flipped94/webook/internal/web"
)

type LoginJWTMiddlewareBuilder struct {
	ignoredPaths []string
}

func NewLoginJWTMiddlewareBuilder() *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{}
}

func (l *LoginJWTMiddlewareBuilder) IgnorePath(path string) *LoginJWTMiddlewareBuilder {
	l.ignoredPaths = append(l.ignoredPaths, path)
	return l
}

func (l *LoginJWTMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		for _, path := range l.ignoredPaths {
			if ctx.Request.URL.Path == path {
				return
			}
		}
		tokenHeader := ctx.GetHeader("Authorization")
		if tokenHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusOK, web.Result{
				Code: http.StatusUnauthorized,
				Msg:  "未登录",
			})
			return
		}
		segs := strings.SplitN(tokenHeader, " ", 2)
		if len(segs) != 2 {
			ctx.AbortWithStatusJSON(http.StatusOK, web.Result{
				Code: http.StatusUnauthorized,
				Msg:  "未登录",
			})
			return
		}
		tokenStr := segs[1]
		claims := &web.UserClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte("O0GWcsczOJHHM8Pu6l2JD9ftliO4Xfou"), nil
		})
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusOK, web.Result{
				Code: http.StatusUnauthorized,
				Msg:  "未登录",
			})
			return
		}
		if token == nil || !token.Valid || claims.Uid == 0 {
			ctx.AbortWithStatusJSON(http.StatusOK, web.Result{
				Code: http.StatusUnauthorized,
				Msg:  "未登录",
			})
			return
		}
		now := time.Now()
		if claims.ExpiresAt.Sub(now) < time.Hour/2 {
			claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Hour))
			token = jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
			tokenStr, _ = token.SignedString([]byte("O0GWcsczOJHHM8Pu6l2JD9ftliO4Xfou"))
			ctx.Header("x-jwt-token", tokenStr)
		}
		ctx.Set("userId", claims.Uid)
	}
}
