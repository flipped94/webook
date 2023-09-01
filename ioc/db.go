package ioc

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/flipped94/webook/config"
)

func InitDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	if err != nil {
		panic(err)
	}
	return db
}
