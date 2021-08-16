package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"shop_srvs/user_srv/global"
	"shop_srvs/user_srv/handler"
	"shop_srvs/user_srv/initialize"
	"shop_srvs/user_srv/proto"
	"shop_srvs/user_srv/utils"
	"syscall"

	uuid "github.com/satori/go.uuid"

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
	Port := flag.Int("port", 0, "端口")
	flag.Parse()

	// 初始化
	initialize.InitLogger()
	initialize.InitConfig()
	initialize.InitDB()

	// 如果命令行没有传递port使用动态端口号，如果传递则使用命令行传递端口号
	if *Port == 0 {
		*Port, _ = utils.GetFreePort()
	}

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
		GRPC:                           fmt.Sprintf("%s:%d", "192.168.0.102", *Port),
		Timeout:                        "5s",
		Interval:                       "5s",
		DeregisterCriticalServiceAfter: "15s",
	}

	// 创建服务注册对象
	registration := new(api.AgentServiceRegistration)
	registration.Name = global.ServiceConfig.Name
	serviceID := fmt.Sprintf("%s", uuid.NewV4())
	//registration.ID = fmt.Sprintf("%s:%d", global.ServiceConfig.Name, *Port)
	registration.ID = serviceID
	registration.Port = *Port
	registration.Tags = []string{"shop", "user_srv"}
	registration.Address = "192.168.0.102"
	registration.Check = check

	err = client.Agent().ServiceRegister(registration)
	if err != nil {
		panic(err)
	}

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
	if err = client.Agent().ServiceDeregister(serviceID); err != nil {
		zap.S().Info("注销失败")
	}
	zap.S().Info("注销成功")

}
