package main

import (
	"flag"
	"fmt"
	"net"
	"shop_srvs/user_srv/handler"
	"shop_srvs/user_srv/initialize"
	"shop_srvs/user_srv/proto"

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
	err = server.Serve(lis)
	if err != nil {
		panic("failed to start grpc:" + err.Error())
	}
}
