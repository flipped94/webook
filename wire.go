//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	"github.com/flipped94/webook/internal/repository"
	"github.com/flipped94/webook/internal/repository/cache/memory"
	"github.com/flipped94/webook/internal/repository/cache/redis"
	"github.com/flipped94/webook/internal/repository/dao"
	"github.com/flipped94/webook/internal/service"
	"github.com/flipped94/webook/internal/web"
	"github.com/flipped94/webook/ioc"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		ioc.InitDB, ioc.InitRedis, ioc.InitGoCache,

		dao.NewUserDao,

		redis.NewUserCache,
		memory.NewCodeCache,

		repository.NewUserRepository,
		repository.NewCodeRepository,

		service.NewUserService,
		service.NewCodeService,

		ioc.InitSmsService,
		web.NewUserHandler,

		ioc.InitWebServer,
		ioc.InitMiddlewares,
	)
	return new(gin.Engine)
}
