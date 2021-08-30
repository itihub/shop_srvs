package main

import (
	"context"
	"fmt"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
)

type OrderListener struct {
}

// 执行本地事务
func (o *OrderListener) ExecuteLocalTransaction(msg *primitive.Message) primitive.LocalTransactionState {
	fmt.Println("开始执行本地事务")
	time.Sleep(time.Second * 3)
	fmt.Println("完成执行本地事务")

	// 本地执行逻辑无缘无故失败 如：代码异常 宕机
	return primitive.UnknowState
}

// 检查本地事务状态 (ExecuteLocalTransaction方法返回非CommitMessageState状态才会执行回查方法)
func (o *OrderListener) CheckLocalTransaction(msg *primitive.MessageExt) primitive.LocalTransactionState {
	fmt.Println("开始回查本地事务状态")
	time.Sleep(time.Second * 10)
	fmt.Println("完成回查本地事务状态")

	// 返回回查结果
	return primitive.CommitMessageState
}

/*
	使用rocketmq 发送事务消息

	测试场景1：
	执行 ExecuteLocalTransaction() 返回 CommitMessageState 模拟本地执行成功
	结果 消息投递成功
	测试场景2：
	执行 ExecuteLocalTransaction() 返回 RollbackMessageState 模拟本地执行失败
	结果 消息丢弃
	测试场景3：
	执行 ExecuteLocalTransaction() 返回 UnknowState 模拟本地执行逻辑无缘无故失败 如：代码异常 宕机
	结果 消息丢弃
*/
func main() {

	p, err := rocketmq.NewTransactionProducer(
		&OrderListener{}, // 事务监听
		producer.WithNameServer([]string{"192.168.56.110:9876"}),
	)
	if err != nil {
		panic(err)
		panic("生成producer失败")
	}

	if err = p.Start(); err != nil {
		panic(err)
		panic("启动producer失败")
	}

	// 发送事务消息
	res, err := p.SendMessageInTransaction(context.Background(), primitive.NewMessage("trans-topic", []byte("tis is transaction message2")))
	if err != nil {
		fmt.Printf("发送失败: %s\n", err.Error())
	} else {
		fmt.Printf("发送成功: %s\n", res.String())
	}

	time.Sleep(time.Hour)
	if err = p.Shutdown(); err != nil {
		panic(err)
		panic("关闭producer失败")
	}
}
