package global

import (
	"shop_srvs/user_srv/config"

	"gorm.io/gorm"
)

// 全局变量
var (
	DB            *gorm.DB
	ServiceConfig config.ServiceConfig
)
