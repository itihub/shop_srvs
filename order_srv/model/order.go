package model

import "time"

// 购物车表
type ShoppingCart struct {
	BaseModel
	User    int32 `gorm:"index;not null;type:int comment '用户ID'"`
	Goods   int32 `gorm:"index;not null;type:int comment '商品ID'"`
	Nums    int32 `gorm:"not null;type:int comment '购买数量'"`
	Checked bool  `gorm:"comment '是否选中'"`
}

// 订单表
type OrderInfo struct {
	BaseModel

	User    int32  `gorm:"index;not null;type:int comment '用户ID'"`
	OrderSn string `gorm:"index;not null;type:varchar(30) comment '订单号'"`                   // 订单号
	PayType string `gorm:"index;not null;type:varchar(20) comment '支付方式 alipay,wechatpay'"` // 订单号

	Status      string     `gorm:"not null;type:varchar(20) comment 'PAYING(待支付),TRADE_SUCCESS(成功),TRADE_CLOSED(超时关闭),WAIT_BUYER_PAY(交易创建),TRADE_FINISHED(交易结束)'"`
	TradeNo     string     `gorm:"type:varchar(100) comment '交易号'"`
	OrderAmount float32    `gorm:"type:float comment '交易金额'"`
	PayTime     *time.Time `gorm:"type:datetime"`

	Address      string `gorm:"type:varchar(100) comment '收件人地址'"`
	SignerName   string `gorm:"type:varchar(20) comment '收件人名称'"`
	SignerMobile string `gorm:"type:varchar(11) comment '收件人手机号码'"`
	Post         string `gorm:"type:varchar(20) comment '留言信息'"`
}

// 订单商品表
type OrderGoods struct {
	BaseModel

	Order int32 `gorm:"index;not null;type:int comment '订单号'"`
	Goods int32 `gorm:"index;not null;type:int comment '商品ID'"`

	GoodsName  string  `gorm:"index;not null;type:varchar(100) comment '商品名称'"`
	GoodsImage string  `gorm:"type:varchar(200) comment '商品图片'"`
	GoodsPrice float32 `gorm:"type:float comment '商品价格'"`
	Nums       int32   `gorm:"type:int comment '购买数量'"`
}
