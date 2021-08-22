package initialize

import (
	"fmt"
	"shop_srvs/order_srv/global"
	"shop_srvs/order_srv/proto"

	_ "github.com/mbobakov/grpc-consul-resolver" // Is’s important
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func InitSrvConn() {
	goodsConn, err := grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?wait=14s", global.ServerConfig.ConsulInfo.Host,
			global.ServerConfig.ConsulInfo.Port, global.ServerConfig.GoodsSrvInfo.Name),
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)
	if err != nil {
		zap.S().Fatal("[InitSrvConn] 连接 【商品服务】 失败")
	}

	global.GoodsSrvClient = proto.NewGoodsClient(goodsConn)

	invConn, err := grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?wait=14s", global.ServerConfig.ConsulInfo.Host,
			global.ServerConfig.ConsulInfo.Port, global.ServerConfig.InventorySrvInfo.Name),
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)
	if err != nil {
		zap.S().Fatal("[InitSrvConn] 连接 【库存服务】 失败")
	}

	global.InventorySrvClient = proto.NewInventoryClient(invConn)
}
