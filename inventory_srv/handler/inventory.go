package handler

import (
	"context"
	"fmt"
	"shop_srvs/inventory_srv/global"
	"shop_srvs/inventory_srv/model"
	"shop_srvs/inventory_srv/proto"
	"sync"

	"github.com/go-redsync/redsync/v4"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type InventoryServer struct {
	proto.UnimplementedInventoryServer // 组合默认实现 防止接口未完全实现启动报错
}

func (s *InventoryServer) SetInv(ctx context.Context, req *proto.GoodsInvInfo) (*emptypb.Empty, error) {
	var inv model.Inventory
	global.DB.Where(&model.Inventory{Goods: req.GoodsId}).First(&inv)
	// 如果goods是0 那么是首次添加进行新增操作
	inv.Goods = req.GoodsId
	inv.Stocks = req.Num
	global.DB.Save(&inv)

	return &emptypb.Empty{}, nil
}

func (s *InventoryServer) InvDetail(ctx context.Context, req *proto.GoodsInvInfo) (*proto.GoodsInvInfo, error) {
	var inv model.Inventory
	if result := global.DB.Where(&model.Inventory{Goods: req.GoodsId}).First(&inv); result.RowsAffected == 0 {
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

	rs := redsync.New(global.RedisPool)

	// 扣减库存
	for _, goodInfo := range req.GoodsInfo {
		// 扣减条件判断
		var inv model.Inventory

		// 获取分布式锁
		mutex := rs.NewMutex(fmt.Sprintf("goods_%d", goodInfo.GoodsId))
		if err := mutex.Lock(); err != nil {
			return nil, status.Errorf(codes.Internal, "获取redis分布式锁异常")
		}

		if result := tx.Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
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
		if ok, err := mutex.Unlock(); !ok || err != nil {
			return nil, status.Errorf(codes.Internal, "释放redis分布式锁异常")
		}
	}

	tx.Commit() // 手动提交
	return &emptypb.Empty{}, nil
}

var m sync.Mutex // 操作系统提供的锁 缺点：在集群部署下竞争同一个数据库时这在单机起作用

//func (s *InventoryServer) Sell(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {
//
//	/*
//		事物
//		扣减库存，进行一批商品库存扣减要么全部成功要么全部失败。本地事物
//		数据库基本的一个应用场景：数据库事物
//	*/
//
//	/*
//		并发
//		并发情况下，可能出现超卖
//	*/
//
//	// 解决方式一
//	//m.Lock() // 获取锁 所有商品扣减库存都抢占一把锁，锁的范围太大 有性能能问题
//
//	tx := global.DB.Begin() // 手动开启事物
//
//	// 扣减库存
//	for _, goodInfo := range req.GoodsInfo {
//		// 扣减条件判断
//		var inv model.Inventory
//
//		// 解决方式二 使用数据库for update 悲观锁
//		//if result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
//
//		for { // 解决方案三
//			if result := tx.Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
//				tx.Rollback() // 回滚之前的操作
//				return nil, status.Errorf(codes.InvalidArgument, "库存信息不存在")
//			}
//			if inv.Stocks < goodInfo.Num {
//				tx.Rollback() // 回滚之前的操作
//				return nil, status.Errorf(codes.ResourceExhausted, "库存不足")
//			}
//			// 进行扣减
//			inv.Stocks -= goodInfo.Num
//
//			// 解决方案三  使用乐观锁机制
//			// update inventory set stocks = stocks - 1, version = version + 1 where goods = goods and version = version
//			// 零值 对于int类型来说 默认值是0 这种会被gorm给忽略掉。 业务已经把库存扣减为0了但是无法更新到数据库中，解决方式 使用更新选定字段
//			if result := tx.Model(&model.Inventory{}).Select("stocks", "version").
//				Where("goods = ? and version = ?", goodInfo.GoodsId, inv.Version).
//				Updates(model.Inventory{Stocks: inv.Stocks, Version: inv.Version + 1}); result.RowsAffected == 0 {
//				zap.S().Info("库存扣减失败")
//			} else {
//				break
//			}
//			//tx.Save(&inv)
//		}
//
//	}
//
//	tx.Commit() // 手动提交
//	//m.Unlock()  // 释放锁
//	return &emptypb.Empty{}, nil
//}

func (s *InventoryServer) Reback(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {
	// 库存归还：1.订单超时归还 2.订单创建失败，归还之前扣减的库存 3. 手动归还

	tx := global.DB.Begin() // 手动开启事物

	for _, goodInfo := range req.GoodsInfo {
		// 扣减条件判断
		var inv model.Inventory
		if result := global.DB.Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
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

/*
	TCC
*/

// tcc 预减库存
func (s *InventoryServer) TrySell(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {

	tx := global.DB.Begin() // 手动开启事物

	rs := redsync.New(global.RedisPool)

	// 扣减库存
	for _, goodInfo := range req.GoodsInfo {
		// 扣减条件判断
		var inv model.InventoryTcc

		// 获取分布式锁
		mutex := rs.NewMutex(fmt.Sprintf("goods_%d", goodInfo.GoodsId))
		if err := mutex.Lock(); err != nil {
			return nil, status.Errorf(codes.Internal, "获取redis分布式锁异常")
		}

		if result := tx.Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
			tx.Rollback() // 回滚之前的操作
			return nil, status.Errorf(codes.InvalidArgument, "库存信息不存在")
		}
		if inv.Stocks < goodInfo.Num {
			tx.Rollback() // 回滚之前的操作
			return nil, status.Errorf(codes.ResourceExhausted, "库存不足")
		}

		// 冻结库存
		inv.Freeze += goodInfo.Num

		tx.Save(&inv)
		if ok, err := mutex.Unlock(); !ok || err != nil {
			return nil, status.Errorf(codes.Internal, "释放redis分布式锁异常")
		}
	}

	tx.Commit() // 手动提交
	return &emptypb.Empty{}, nil
}

// tcc 确认库存
func (s *InventoryServer) ConfirmSell(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {

	tx := global.DB.Begin() // 手动开启事物

	rs := redsync.New(global.RedisPool)

	// 扣减库存
	for _, goodInfo := range req.GoodsInfo {
		// 扣减条件判断
		var inv model.InventoryTcc

		// 获取分布式锁
		mutex := rs.NewMutex(fmt.Sprintf("goods_%d", goodInfo.GoodsId))
		if err := mutex.Lock(); err != nil {
			return nil, status.Errorf(codes.Internal, "获取redis分布式锁异常")
		}

		if result := tx.Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
			tx.Rollback() // 回滚之前的操作
			return nil, status.Errorf(codes.InvalidArgument, "库存信息不存在")
		}
		if inv.Stocks < goodInfo.Num {
			tx.Rollback() // 回滚之前的操作
			return nil, status.Errorf(codes.ResourceExhausted, "库存不足")
		}

		// 冻结库存释放
		inv.Freeze -= goodInfo.Num
		// 真正扣减库存
		inv.Stocks -= goodInfo.Num

		tx.Save(&inv)
		if ok, err := mutex.Unlock(); !ok || err != nil {
			return nil, status.Errorf(codes.Internal, "释放redis分布式锁异常")
		}
	}

	tx.Commit() // 手动提交
	return &emptypb.Empty{}, nil
}

// tcc 回滚库存
func (s *InventoryServer) CancelSell(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {

	tx := global.DB.Begin() // 手动开启事物

	rs := redsync.New(global.RedisPool)

	// 扣减库存
	for _, goodInfo := range req.GoodsInfo {
		// 扣减条件判断
		var inv model.InventoryTcc

		// 获取分布式锁
		mutex := rs.NewMutex(fmt.Sprintf("goods_%d", goodInfo.GoodsId))
		if err := mutex.Lock(); err != nil {
			return nil, status.Errorf(codes.Internal, "获取redis分布式锁异常")
		}

		if result := tx.Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
			tx.Rollback() // 回滚之前的操作
			return nil, status.Errorf(codes.InvalidArgument, "库存信息不存在")
		}
		if inv.Stocks < goodInfo.Num {
			tx.Rollback() // 回滚之前的操作
			return nil, status.Errorf(codes.ResourceExhausted, "库存不足")
		}

		// 冻结库存释放
		inv.Freeze -= goodInfo.Num

		tx.Save(&inv)
		if ok, err := mutex.Unlock(); !ok || err != nil {
			return nil, status.Errorf(codes.Internal, "释放redis分布式锁异常")
		}
	}

	tx.Commit() // 手动提交
	return &emptypb.Empty{}, nil
}

func (s *InventoryServer) InvDetailTcc(ctx context.Context, req *proto.GoodsInvInfo) (*proto.GoodsInvInfo, error) {
	var inv model.InventoryTcc
	if result := global.DB.Where(&model.Inventory{Goods: req.GoodsId}).First(&inv); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "库存信息不存在")
	}

	return &proto.GoodsInvInfo{
		GoodsId: inv.Goods,
		Num:     inv.Stocks - inv.Freeze,
	}, nil
}
