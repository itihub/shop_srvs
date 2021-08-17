package tests

import (
	"context"
	"fmt"
	"shop_srvs/goods_srv/proto"

	"google.golang.org/grpc"
)

var (
	goodsClient proto.GoodsClient
	conn        *grpc.ClientConn
)

func Init() {
	var err error
	conn, err = grpc.Dial(fmt.Sprintf("127.0.0.1:%d", 50051), grpc.WithInsecure()) // 拨号 建立连接
	if err != nil {
		panic(err)
	}

	goodsClient = proto.NewGoodsClient(conn) // 生成客户端
}

func main() {
	Init()

	TestGetBrandList()

	conn.Close() // 关闭连接

}

func TestGetBrandList() {
	rsp, err := goodsClient.BrandList(context.Background(), &proto.BrandFilterRequest{})
	if err != nil {
		panic(err)
	}

	for _, brand := range rsp.Data {
		fmt.Println(brand.Name)
	}
}
