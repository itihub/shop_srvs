package main

import (
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"net"
	"shop_srvs/user_srv/handler"
	"shop_srvs/user_srv/proto"
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
	fmt.Println("ip: ", *IP)
	fmt.Println("port: ", *Port)

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
