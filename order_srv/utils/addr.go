package utils

import (
	"fmt"
	"net"
)

// 动态获取可用端口
func GetFreePort() (int, error) {

	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, nil
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, nil
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func main() {
	port, _ := GetFreePort()
	fmt.Println(port)
}
