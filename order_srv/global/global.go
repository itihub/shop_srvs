package global

import (
	"shop_srvs/order_srv/config"
	"shop_srvs/order_srv/proto"

	"github.com/go-redsync/redsync/v4/redis"

	"gorm.io/gorm"
)

// 全局变量
var (
	DB                 *gorm.DB
	ServerConfig       config.ServerConfig
	NacosConfig        config.NacosConfig
	RedisPool          redis.Pool
	GoodsSrvClient     proto.GoodsClient
	InventorySrvClient proto.InventoryClient
)
