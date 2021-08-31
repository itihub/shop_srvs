package global

import (
	"shop_srvs/userop_srv/config"

	"github.com/go-redsync/redsync/v4/redis"

	"gorm.io/gorm"
)

// 全局变量
var (
	DB           *gorm.DB
	ServerConfig config.ServerConfig
	NacosConfig  config.NacosConfig
	RedisPool    redis.Pool
)
