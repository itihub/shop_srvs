package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"shop_srvs/order_srv/global"
	"shop_srvs/order_srv/model"
	"shop_srvs/order_srv/proto"
	"time"

	"github.com/opentracing/opentracing-go"

	"github.com/apache/rocketmq-client-go/v2/consumer"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"go.uber.org/zap"

	"google.golang.org/protobuf/types/known/emptypb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderService struct {
	proto.UnimplementedOrderServer
}

func (s *OrderService) CartItemList(ctx context.Context, req *proto.UserInfo) (*proto.CartItemListResponse, error) {
	// 获取用户的购物车列表
	var shopCarts []model.ShoppingCart
	var rsp proto.CartItemListResponse

	if result := global.DB.Where(&model.ShoppingCart{User: req.Id}).Find(&shopCarts); result.Error != nil {
		return nil, status.Errorf(codes.Unimplemented, result.Error.Error())
	} else {
		rsp = proto.CartItemListResponse{
			Total: int32(result.RowsAffected),
		}
	}

	for _, shopCart := range shopCarts {
		rsp.Data = append(rsp.Data, &proto.ShopCartInfoResponse{
			Id:      shopCart.ID,
			UserId:  shopCart.User,
			GoodsId: shopCart.Goods,
			Nums:    shopCart.Nums,
			Checked: shopCart.Checked,
		})
	}

	return &rsp, nil

}
func (s *OrderService) CreateCartItem(ctx context.Context, req *proto.CartItemRequest) (*proto.ShopCartInfoResponse, error) {
	// 将商品到购物车

	var shopCart model.ShoppingCart

	// 1. 购物车存在该商品 进行合并
	if result := global.DB.Where(&model.ShoppingCart{Goods: req.GoodsId, User: req.UserId}).First(&shopCart); result.RowsAffected == 1 {
		shopCart.Nums += req.Nums
	} else {
		// 2. 购物车不存在该商品
		shopCart.User = req.UserId
		shopCart.Goods = req.GoodsId
		shopCart.Nums = req.Nums
		shopCart.Checked = false
	}
	global.DB.Save(&shopCart)

	return &proto.ShopCartInfoResponse{Id: shopCart.ID}, nil
}

func (s *OrderService) UpdateCartItem(ctx context.Context, req *proto.CartItemRequest) (*emptypb.Empty, error) {
	// 更新购物车记录 更新数量和选择状态
	var shopCart model.ShoppingCart

	if result := global.DB.Where("goods = ? and user = ?", req.GoodsId, req.UserId).First(&shopCart); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "记录不存在")
	}

	shopCart.Checked = req.Checked
	if req.Nums > 0 {
		shopCart.Nums = req.Nums
	}
	global.DB.Save(&shopCart)

	return &emptypb.Empty{}, nil
}

func (s *OrderService) DeleteCartItem(ctx context.Context, req *proto.CartItemRequest) (*emptypb.Empty, error) {
	if result := global.DB.Where("goods = ? and user = ?", req.GoodsId, req.UserId).Delete(&model.ShoppingCart{}); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "购物车记录不存在")
	}

	return &emptypb.Empty{}, nil
}

func (s *OrderService) CreateOrder(ctx context.Context, req *proto.OrderRequest) (*proto.OrderInfoResponse, error) {
	/*
		新建订单
			1. 从购物车中获取选中的商品
			2. 商品的金额 - 访问商品服务 （跨服务调用）
			3. 库存扣减 - 访问库存服务 （跨服务调用）
			4. 订单的基本信息表 - 订单的商品信息表
			5. 从购物车中删除已购买记录
	*/

	orderListener := OrderListener{Ctx: ctx}
	p, err := rocketmq.NewTransactionProducer(
		&orderListener, // 事务监听
		producer.WithNameServer([]string{fmt.Sprintf("%s:%d", global.ServerConfig.RocketMQInfo.Host, global.ServerConfig.RocketMQInfo.Port)}),
	)
	if err != nil {
		zap.S().Errorf("生成producer失败: %s", err.Error())
		return nil, err
	}
	if err = p.Start(); err != nil {
		zap.S().Errorf("启动producer失败: %s", err.Error())
		return nil, err
	}

	order := model.OrderInfo{
		User:         req.UserId,
		OrderSn:      GenerateOrderSn(req.UserId),
		Address:      req.Address,
		SignerName:   req.Name,
		SignerMobile: req.Mobile,
		Post:         req.Post,
	}
	jsonString, _ := json.Marshal(order)

	// 发送归还事务消息
	_, err = p.SendMessageInTransaction(context.Background(), primitive.NewMessage(global.ServerConfig.RocketMQInfo.OrderRebackTopic, jsonString))
	if err != nil {
		zap.S().Errorf("")
		return nil, status.Error(codes.Internal, "发送消息失败")
	}
	if orderListener.Code != codes.OK {
		return nil, status.Errorf(codes.Internal, orderListener.Detail)
	}

	return &proto.OrderInfoResponse{
		Id:      orderListener.ID,
		OrderSn: order.OrderSn,
		Total:   orderListener.OrderAmount,
	}, nil

}

type OrderListener struct {
	Code        codes.Code
	Detail      string
	ID          int32
	OrderAmount float32
	Ctx         context.Context
}

// 执行本地事务
func (o *OrderListener) ExecuteLocalTransaction(msg *primitive.Message) primitive.LocalTransactionState {

	// 反序列化msg
	var orderInfo model.OrderInfo
	json.Unmarshal(msg.Body, &orderInfo)

	parentSpan := opentracing.SpanFromContext(o.Ctx)

	var goodsIds []int32
	var shopCarts []model.ShoppingCart
	goodsNumsMap := make(map[int32]int32) // key is goodsId value is nums
	shopCartSpan := opentracing.GlobalTracer().StartSpan("select_shop_cart", opentracing.ChildOf(parentSpan.Context()))
	if result := global.DB.Where(&model.ShoppingCart{User: orderInfo.User, Checked: true}).Find(&shopCarts); result.RowsAffected == 0 {
		o.Code = codes.InvalidArgument
		o.Detail = "没有选中结算的商品"
		return primitive.RollbackMessageState
	}
	shopCartSpan.Finish()

	for _, shopCart := range shopCarts {
		goodsIds = append(goodsIds, shopCart.Goods)
		goodsNumsMap[shopCart.Goods] = shopCart.Nums
	}

	// 跨服务调用 - 批量查询商品
	queryGoodsSpan := opentracing.GlobalTracer().StartSpan("query_goods", opentracing.ChildOf(parentSpan.Context()))
	goods, err := global.GoodsSrvClient.BatchGetGoods(context.Background(), &proto.BatchGoodsIdInfo{Id: goodsIds})
	if err != nil {
		o.Code = codes.Internal
		o.Detail = "批量查询商品信息失败"
		return primitive.RollbackMessageState
	}
	queryGoodsSpan.Finish()
	var orderAmount float32
	var orderGoods []*model.OrderGoods
	var goodsInvInfo []*proto.GoodsInvInfo
	for _, good := range goods.Data {
		orderAmount += good.ShopPrice * float32(goodsNumsMap[good.Id])
		orderGoods = append(orderGoods, &model.OrderGoods{
			Goods:      good.Id,
			GoodsName:  good.Name,
			GoodsImage: good.GoodsFrontImage,
			GoodsPrice: good.ShopPrice,
			Nums:       goodsNumsMap[good.Id],
		})
		goodsInvInfo = append(goodsInvInfo, &proto.GoodsInvInfo{
			GoodsId: good.Id,
			Num:     goodsNumsMap[good.Id],
		})
	}

	// 跨服务调用 - 库存扣减
	queryInvSpan := opentracing.GlobalTracer().StartSpan("query_inv", opentracing.ChildOf(parentSpan.Context()))
	_, err = global.InventorySrvClient.Sell(context.Background(), &proto.SellInfo{OrderSn: orderInfo.OrderSn, GoodsInfo: goodsInvInfo})
	if err != nil {
		o.Code = codes.ResourceExhausted
		o.Detail = "扣减库存失败"
		return primitive.RollbackMessageState
	}
	queryInvSpan.Finish()

	// 生成订单表
	// 20210308xxxx
	tx := global.DB.Begin()

	orderInfo.OrderAmount = orderAmount

	saveOrderSpan := opentracing.GlobalTracer().StartSpan("save_order", opentracing.ChildOf(parentSpan.Context()))
	if result := tx.Save(&orderInfo); result.RowsAffected == 0 {
		tx.Rollback()
		o.Code = codes.Internal
		o.Detail = "创建订单失败"
		return primitive.CommitMessageState
	}
	saveOrderSpan.Finish()

	o.OrderAmount = orderAmount
	o.ID = orderInfo.ID

	for _, orderGood := range orderGoods {
		orderGood.Order = orderInfo.ID
	}

	// 批量插入orderGoods
	saveOrderGoodsSpan := opentracing.GlobalTracer().StartSpan("save_order_goods", opentracing.ChildOf(parentSpan.Context()))
	if result := tx.CreateInBatches(&orderGoods, 100); result.RowsAffected == 0 {
		tx.Rollback()
		o.Code = codes.Internal
		o.Detail = "批量插入订单商品失败"
		return primitive.CommitMessageState
	}
	saveOrderGoodsSpan.Finish()

	deleteShopCartSpan := opentracing.GlobalTracer().StartSpan("delete_shop_cart", opentracing.ChildOf(parentSpan.Context()))
	if result := tx.Where(&model.ShoppingCart{User: orderInfo.User, Checked: true}).Delete(&model.ShoppingCart{}); result.RowsAffected == 0 {
		tx.Rollback()
		o.Code = codes.Internal
		o.Detail = "删除购物车记录失败"
		return primitive.CommitMessageState
	}
	deleteShopCartSpan.Finish()

	sendOrderTimeoutMsgSpan := opentracing.GlobalTracer().StartSpan("send_order_timeout_msg", opentracing.ChildOf(parentSpan.Context()))
	// 发送未支付归还库存延时消息
	p, err := rocketmq.NewProducer(producer.WithNameServer([]string{fmt.Sprintf("%s:%d", global.ServerConfig.RocketMQInfo.Host, global.ServerConfig.RocketMQInfo.Port)}))
	if err != nil {
		panic(err)
		panic("生成producer失败")
	}

	if err = p.Start(); err != nil {
		panic(err)
		panic("启动producer失败")
	}

	// 延迟消息
	msg = primitive.NewMessage(global.ServerConfig.RocketMQInfo.OrderTimeoutTopic, msg.Body)
	msg.WithDelayTimeLevel(5)

	// 同步发送
	_, err = p.SendSync(context.Background(), msg)
	if err != nil {
		zap.S().Errorf("发送延迟消息失败: %v\n", err)
		tx.Rollback()
		o.Code = codes.Internal
		o.Detail = "发送延迟消息失败"
		return primitive.CommitMessageState
	}
	sendOrderTimeoutMsgSpan.Finish()

	tx.Commit()
	o.Code = codes.OK
	return primitive.RollbackMessageState
}

// 检查本地事务状态 (ExecuteLocalTransaction方法返回非CommitMessageState状态才会执行回查方法)
func (o *OrderListener) CheckLocalTransaction(msg *primitive.MessageExt) primitive.LocalTransactionState {
	// 反序列化msg
	var orderInfo model.OrderInfo
	json.Unmarshal(msg.Body, &orderInfo)

	parentSpan := opentracing.SpanFromContext(o.Ctx)
	checkOrderExistSpan := opentracing.GlobalTracer().StartSpan("check_order_exist", opentracing.ChildOf(parentSpan.Context()))
	// 检查订单是否存在
	if result := global.DB.Where(model.OrderInfo{OrderSn: orderInfo.OrderSn}).First(&orderInfo); result.RowsAffected == 0 {
		// 订单不存在，并不能说明库存已经被扣减
		return primitive.CommitMessageState
	}
	checkOrderExistSpan.Finish()

	return primitive.RollbackMessageState
}

func (s *OrderService) OrderList(ctx context.Context, req *proto.OrderFilterRequest) (*proto.OrderListResponse, error) {
	var orders []model.OrderInfo
	var rsp proto.OrderListResponse

	// 后台管理系统查询 还是 电商系统查询
	var total int64
	global.DB.Where(&model.OrderInfo{User: req.UserId}).Count(&total)
	rsp.Total = int32(total)

	global.DB.Scopes(Paginate(int(req.Pages), int(req.PagePerNums))).Find(&orders)
	for _, order := range orders {
		rsp.Data = append(rsp.Data, &proto.OrderInfoResponse{
			Id:      order.ID,
			UserId:  order.User,
			OrderSn: order.OrderSn,
			PayType: order.PayType,
			Status:  order.Status,
			Post:    order.Post,
			Total:   order.OrderAmount,
			Address: order.Address,
			Name:    order.SignerName,
			Mobile:  order.SignerMobile,
			AddTime: order.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return &rsp, nil
}
func (s *OrderService) OrderDetail(ctx context.Context, req *proto.OrderRequest) (*proto.OrderInfoDetailResponse, error) {
	var order model.OrderInfo
	var rsp proto.OrderInfoDetailResponse

	// 该订单的ID是否是当前用户的订单，如果web层用户传递过来一个ID的订单，web层应该先查询订单ID是否是当前用户的订单
	// 在个人中心可以这样做，但是如果在后台管理系统，web层如果是后台管理系统，那么只传递order的ID，如果是电商心态还需要用户ID
	if result := global.DB.Where(&model.OrderInfo{User: req.UserId, BaseModel: model.BaseModel{ID: req.Id}}).First(&order); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "订单不存在")
	}

	orderInfo := proto.OrderInfoResponse{}

	orderInfo.Id = order.ID
	orderInfo.UserId = order.User
	orderInfo.OrderSn = order.OrderSn
	orderInfo.PayType = order.PayType
	orderInfo.Status = order.Status
	orderInfo.Post = order.Post
	orderInfo.Total = order.OrderAmount
	orderInfo.Address = order.Address
	orderInfo.Name = order.SignerName
	orderInfo.Mobile = order.SignerMobile

	rsp.OrderInfo = &orderInfo

	var orderGoods []model.OrderGoods
	global.DB.Where(&model.OrderGoods{Order: order.ID}).Find(&orderGoods)
	for _, orderGood := range orderGoods {
		rsp.Goods = append(rsp.Goods, &proto.OrderItemResponse{
			GoodsId:    orderGood.Goods,
			GoodsName:  orderGood.GoodsName,
			GoodsPrice: orderGood.GoodsPrice,
			Nums:       orderGood.Nums,
		})
	}

	return &rsp, nil
}
func (s *OrderService) UpdateOrderStatus(ctx context.Context, req *proto.OrderStatus) (*emptypb.Empty, error) {
	// 方式一：先查询，再更新 实际上有两条sql执行 select 和 update 语句
	// 方式二：没有过多的业务逻辑，直接执行update语句 通过判断rows来决定是否成功
	if result := global.DB.Model(&model.OrderInfo{}).Where("order_sn = ?", req.OrderSn).Update("status", req.Status); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "订单不存在")
	}
	return &emptypb.Empty{}, nil
}

func GenerateOrderSn(userId int32) string {
	/*
		订单号生成规则
		年月日时分秒+用户ID+2位随机数
	*/
	now := time.Now()
	rand.Seed(time.Now().UnixNano()) // 设置随机数种子
	orderSn := fmt.Sprintf("%d%d%d%d%d%d%d%d",
		now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(),
		userId, rand.Intn(90)+10)
	return orderSn
}

func OrderTimeout(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	for i := range msgs {
		// 反序列化msg
		var orderInfo model.OrderInfo
		json.Unmarshal(msgs[i].Body, &orderInfo)

		// 查询订单支付状态 已支付什么都不做  未支付归还库存
		var order model.OrderInfo
		if result := global.DB.Model(model.OrderInfo{}).Where(model.OrderInfo{OrderSn: orderInfo.OrderSn}).First(&order); result.RowsAffected == 0 {
			return consumer.ConsumeSuccess, nil
		}

		if order.Status != "TRADE_SUCCESS" {

			tx := global.DB.Begin()

			// 修改订单状态为已支付
			order.Status = "TRADE_CLOSED"
			tx.Save(&order)

			// 归还库存 模仿order中发送一个归还库存的事务消息
			p, err := rocketmq.NewProducer(producer.WithNameServer([]string{fmt.Sprintf("%s:%d", global.ServerConfig.RocketMQInfo.Host, global.ServerConfig.RocketMQInfo.Port)}))
			if err != nil {
				panic(err)
				panic("生成producer失败")
			}

			if err = p.Start(); err != nil {
				panic(err)
				panic("启动producer失败")
			}

			// 同步发送
			_, err = p.SendSync(context.Background(), primitive.NewMessage(global.ServerConfig.RocketMQInfo.OrderRebackTopic, msgs[i].Body))
			if err != nil {
				tx.Rollback()
				fmt.Printf("发送失败: %s\n", err.Error())
				return consumer.ConsumeRetryLater, nil
			}
		}
	}
	return consumer.ConsumeSuccess, nil
}
