//go:build !k8s

package config

var Config = config{
	DB: DBConfig{
		DSN: "root:123456@tcp(192.168.137.133:3306)/webook",
	},
	Redis: RedisConfig{
		Addr: "192.168.137.133:6379",
	},
}
