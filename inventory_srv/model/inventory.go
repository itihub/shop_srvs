package model

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
	user   int32 `gorm:"not null;index;type:int comment '用户ID'"`
	goods  int32 `gorm:"not null;index;type:int comment '商品ID'"`
	nums   int32 `gorm:"not null;type:int comment '购买数量'"`
	order  int32 `gorm:"not null;index;type:int comment '订单号'"`
	status int32 `gorm:"not null;type:int comment '状态'"` // 1. 表示库存是预减，幂等性 2. 表示已支付
}

// tcc
type InventoryTcc struct {
	BaseModel
	Goods  int32 `gorm:"not null;index;type:int comment '商品ID'"`
	Stocks int32 `gorm:"not null;type:int comment '库存'"`
	Freeze int32 `gorm:"not null;type:int comment '冻结库存'"`
}
