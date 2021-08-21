package global

import (
	"shop_srvs/inventory_srv/config"

	"github.com/go-redsync/redsync/v4/redis"

	"gorm.io/gorm"
)

// 全局变量
var (
	DB            *gorm.DB
	ServiceConfig config.ServiceConfig
	NacosConfig   config.NacosConfig
	RedisPool     redis.Pool
)
