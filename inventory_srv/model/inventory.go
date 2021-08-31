package model

import (
	"database/sql/driver"
	"encoding/json"
)

// 仓库表
//type Stock struct {
//	BaseModel
//	Name string
//	Address string
//}

type Inventory struct {
	BaseModel
	Goods   int32 `gorm:"not null;index;type:int comment '商品ID'"`
	Stocks  int32 `gorm:"not null;type:int comment '库存'"`
	Version int32 `gorm:"not null;type:int comment '分布锁'"` // 乐观锁
}

// 扣减库存记录
type InventoryHistory struct {
	User   int32 `gorm:"not null;index;type:int comment '用户ID'"`
	Goods  int32 `gorm:"not null;index;type:int comment '商品ID'"`
	Nums   int32 `gorm:"not null;type:int comment '购买数量'"`
	Order  int32 `gorm:"not null;index;type:int comment '订单号'"`
	Status int32 `gorm:"not null;type:int comment '状态'"` // 1. 表示库存是预减，幂等性 2. 表示已支付
}

type GoodsDetail struct {
	Goods int32
	Num   int32
}

type GoodsDetailList []GoodsDetail

func (g *GoodsDetailList) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), &g)
}

func (g GoodsDetailList) Value() (driver.Value, error) {
	return json.Marshal(g)
}

type StockSellDetail struct {
	OrderSn string          `gorm:"not null;index:idx_order_sn,unique;type:varchar(200) comment '订单号'"`
	Status  int32           `gorm:"not null;index;type:int comment '状态:1:表示已扣减,2:表示已归还'"`
	Detail  GoodsDetailList `gorm:"not null;type:varchar(200) comment '商品详情'"`
}

// tcc
type InventoryTcc struct {
	BaseModel
	Goods  int32 `gorm:"not null;index;type:int comment '商品ID'"`
	Stocks int32 `gorm:"not null;type:int comment '库存'"`
	Freeze int32 `gorm:"not null;type:int comment '冻结库存'"`
}
