package global

import (
	"shop_srvs/goods_srv/config"

	"github.com/olivere/elastic/v7"
	"gorm.io/gorm"
)

// 全局变量
var (
	DB            *gorm.DB
	ServerConfig  config.ServerConfig
	NacosConfig   config.NacosConfig
	ElasticClient *elastic.Client
)
