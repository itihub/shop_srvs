package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"shop_srvs/goods_srv/global"
	"shop_srvs/goods_srv/initialize"
	"shop_srvs/goods_srv/proto"
	"shop_srvs/goods_srv/utils"
	"shop_srvs/goods_srv/utils/register/consul"
	"syscall"

	uuid "github.com/satori/go.uuid"

	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	/*
		go build main.go
		main.exe -h 提示
		main.exe 使用默认参数启动
		main.exe -port 50053  使用输入参数启动
	*/
	// main.exe -h

	// 通过flag获取参数 (命令行输入参数)
	IP := flag.String("ip", "0.0.0.0", "IP地址")
	Port := flag.Int("port", 0, "端口")
	flag.Parse()

	// 初始化
	initialize.InitLogger()
	initialize.InitConfig()
	initialize.InitDB()

	// 如果命令行没有传递port使用动态端口号，如果传递则使用命令行传递端口号
	if *Port == 0 {
		if global.ServiceConfig.Port == 0 {
			*Port, _ = utils.GetFreePort()
		}
		*Port = global.ServiceConfig.Port
	}

	zap.S().Info("ip: ", *IP)
	zap.S().Info("port: ", *Port)

	// GRPC 启动
	server := grpc.NewServer()
	proto.RegisterGoodsServer(server, &proto.UnimplementedGoodsServer{})
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *IP, *Port))
	if err != nil {
		panic("failed to listen:" + err.Error())
	}

	// grpc 注册服务健康检查
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())

	// 服务注册
	serviceID := fmt.Sprintf("%s", uuid.NewV4())
	registryClient := consul.NewRegistryClient(global.ServiceConfig.ConsulInfo.Host, global.ServiceConfig.ConsulInfo.Port)
	err = registryClient.Register(global.ServiceConfig.Host, global.ServiceConfig.Port, global.ServiceConfig.Name, global.ServiceConfig.Tags, serviceID)
	if err != nil {
		zap.S().Panic("服务注册失败:", err.Error())
	}
	zap.S().Debugf("启动服务, 端口:%d", global.ServiceConfig.Port)

	go func() {
		err = server.Serve(lis)
		if err != nil {
			panic("failed to start grpc:" + err.Error())
		}
	}()

	// 接受终止信号 优雅退出
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	if err = registryClient.DeRegister(serviceID); err != nil {
		zap.S().Info("注销失败")
	}
	zap.S().Info("注销成功")

}
