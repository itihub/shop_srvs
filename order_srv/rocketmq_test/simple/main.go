package main

import (
	"context"
	"fmt"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
)

/*
	使用rocketmq 发送简单消息
*/
func main() {

	p, err := rocketmq.NewProducer(producer.WithNameServer([]string{"192.168.56.110:9876"}))
	if err != nil {
		panic(err)
		panic("生成producer失败")
	}

	if err = p.Start(); err != nil {
		panic(err)
		panic("启动producer失败")
	}

	// 同步发送
	res, err := p.SendSync(context.Background(), primitive.NewMessage("shop-test", []byte("tis is shop-test1")))
	if err != nil {
		fmt.Printf("发送失败: %s\n", err.Error())
	} else {
		fmt.Printf("发送成功: %s\n", res.String())
	}

	if err = p.Shutdown(); err != nil {
		panic(err)
		panic("关闭producer失败")
	}
}
