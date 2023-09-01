// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/flipped94/webook/internal/repository"
	"github.com/flipped94/webook/internal/repository/cache/memory"
	"github.com/flipped94/webook/internal/repository/cache/redis"
	"github.com/flipped94/webook/internal/repository/dao"
	"github.com/flipped94/webook/internal/service"
	"github.com/flipped94/webook/internal/web"
	"github.com/flipped94/webook/ioc"
	"github.com/gin-gonic/gin"
)

// Injectors from wire.go:

func InitWebServer() *gin.Engine {
	cmdable := ioc.InitRedis()
	v := ioc.InitMiddlewares(cmdable)
	db := ioc.InitDB()
	userDao := dao.NewUserDao(db)
	userCache := redis.NewUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDao, userCache)
	userService := service.NewUserService(userRepository)
	iCache := ioc.InitGoCache()
	codeCache := memory.NewCodeCache(iCache)
	codeRepository := repository.NewCodeRepository(codeCache)
	smsService := ioc.InitSmsService()
	codeService := service.NewCodeService(codeRepository, smsService)
	userHandler := web.NewUserHandler(userService, codeService)
	engine := ioc.InitWebServer(v, userHandler)
	return engine
}
