package main

import (
	"context"
	"fmt"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

/*
	使用rocketmq client消费消息
*/
func main() {

	// 使用推的方式监听消息
	c, _ := rocketmq.NewPushConsumer(
		consumer.WithNameServer([]string{"192.168.56.110:9876"}),
		consumer.WithGroupName("shop-test-group"), // 设置group 作用：集群部署时，可以起到负载均衡的作用，一个消息只会被消费一次
	)

	// 订阅消息
	if err := c.Subscribe("shop-test", consumer.MessageSelector{}, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {

		for i := range msgs {
			fmt.Printf("获取到值: %v \n", msgs[i])
		}

		return consumer.ConsumeSuccess, nil
	}); err != nil {
		fmt.Println("读取消息失败")
	}

	_ = c.Start()

	// 不能让主goroutine退出
	time.Sleep(time.Hour)
	_ = c.Shutdown()
}
