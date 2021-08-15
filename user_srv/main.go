package main

import (
	"flag"
	"fmt"
	"net"
	"shop_srvs/user_srv/global"
	"shop_srvs/user_srv/handler"
	"shop_srvs/user_srv/initialize"
	"shop_srvs/user_srv/proto"

	"github.com/hashicorp/consul/api"

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
	Port := flag.Int("port", 50051, "端口")
	flag.Parse()

	// 初始化
	initialize.InitLogger()
	initialize.InitConfig()
	initialize.InitDB()

	zap.S().Info("ip: ", *IP)
	zap.S().Info("port: ", *Port)

	// GRPC 启动
	server := grpc.NewServer()
	proto.RegisterUserServer(server, &handler.UserService{})
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *IP, *Port))
	if err != nil {
		panic("failed to listen:" + err.Error())
	}

	// grpc 注册服务健康检查
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())

	// 服务注册
	cfg := api.DefaultConfig()
	cfg.Address = fmt.Sprintf("%s:%d", global.ServiceConfig.ConsulInfo.Host,
		global.ServiceConfig.ConsulInfo.Port)

	client, err := api.NewClient(cfg)
	if err != nil {
		panic(err)
	}

	// 创建服务检查检查对象
	check := &api.AgentServiceCheck{
		GRPC:                           fmt.Sprintf("%s:%d", "172.20.10.2", *Port),
		Timeout:                        "5s",
		Interval:                       "5s",
		DeregisterCriticalServiceAfter: "15s",
	}

	// 创建服务注册对象
	registration := new(api.AgentServiceRegistration)
	registration.Name = global.ServiceConfig.Name
	registration.ID = global.ServiceConfig.Name
	registration.Port = *Port
	registration.Tags = []string{"shop", "user_srv"}
	registration.Address = "172.20.10.2"
	registration.Check = check

	err = client.Agent().ServiceRegister(registration)
	if err != nil {
		panic(err)
	}

	err = server.Serve(lis)
	if err != nil {
		panic("failed to start grpc:" + err.Error())
	}
}
