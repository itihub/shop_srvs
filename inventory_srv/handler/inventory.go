package handler

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"shop_srvs/inventory_srv/global"
	"shop_srvs/inventory_srv/model"
	"shop_srvs/inventory_srv/proto"
)

type InventoryServer struct {
	proto.UnimplementedInventoryServer // 组合默认实现 防止接口未完全实现启动报错
}

func (s *InventoryServer) SetInv(ctx context.Context, req *proto.GoodsInvInfo) (*emptypb.Empty, error) {
	var inv model.Inventory
	global.DB.First(&inv, req.GoodsId)
	// 如果goods是0 那么是首次添加进行新增操作
	inv.Goods = req.GoodsId
	inv.Stocks = req.Num
	global.DB.Save(&inv)

	return &emptypb.Empty{}, nil
}
func (s *InventoryServer) InvDetail(ctx context.Context, req *proto.GoodsInvInfo) (*proto.GoodsInvInfo, error) {
	var inv model.Inventory
	if result := global.DB.First(&inv, req.GoodsId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "库存信息不存在")
	}

	return &proto.GoodsInvInfo{
		GoodsId: inv.Goods,
		Num:     inv.Stocks,
	}, nil
}
func (s *InventoryServer) Sell(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {

	/*
		事物
		扣减库存，进行一批商品库存扣减要么全部成功要么全部失败。本地事物
		数据库基本的一个应用场景：数据库事物
	*/

	/*
		并发
		并发情况下，可能出现超卖
	*/

	tx := global.DB.Begin() // 手动开启事物

	// 扣减库存
	for _, goodInfo := range req.GoodsInfo {
		// 扣减条件判断
		var inv model.Inventory
		if result := global.DB.First(&inv, goodInfo.GoodsId); result.RowsAffected == 0 {
			tx.Rollback() // 回滚之前的操作
			return nil, status.Errorf(codes.InvalidArgument, "库存信息不存在")
		}
		if inv.Stocks < goodInfo.Num {
			tx.Rollback() // 回滚之前的操作
			return nil, status.Errorf(codes.ResourceExhausted, "库存不足")
		}
		// 进行扣减
		inv.Stocks -= goodInfo.Num
		tx.Save(&inv)
	}

	tx.Commit() // 手动提交
	return &emptypb.Empty{}, nil
}
func (s *InventoryServer) Reback(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {
	// 库存归还：1.订单超时归还 2.订单创建失败，归还之前扣减的库存 3. 手动归还

	tx := global.DB.Begin() // 手动开启事物

	for _, goodInfo := range req.GoodsInfo {
		// 扣减条件判断
		var inv model.Inventory
		if result := global.DB.First(&inv, goodInfo.GoodsId); result.RowsAffected == 0 {
			tx.Rollback() // 回滚之前的操作
			return nil, status.Errorf(codes.InvalidArgument, "库存信息不存在")
		}
		// 进行库存归还
		inv.Stocks += goodInfo.Num
		tx.Save(&inv)
	}

	tx.Commit() // 手动提交
	return &emptypb.Empty{}, nil
}
