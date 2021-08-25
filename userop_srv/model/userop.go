package model

const (
	LEAVING_MESSAGES = iota + 1
	COMPLAINT
	INQUIRY
	POST_SALE
	WANT_TO_BUY
)

// 留言表
type LeavingMessages struct {
	BaseModel

	User        int32  `gorm:"index;not null;type:int comment '用户ID'"`
	MessageType int32  `gorm:"not null;type:int comment '留言类型: 1(留言),2(投诉),3(询问),4(售后),5(求购)'"`
	Subject     string `gorm:"type:varchar(100) comment '主题'"`

	Message string `gorm:"type:text comment '内容'"`
	File    string `gorm:"type:varchar(200) comment '文件地址'"`
}

// 地址表
type Address struct {
	BaseModel

	User         int32  `gorm:"index;not null;type:int comment '用户ID'"`
	Province     string `gorm:"type:varchar(10) comment '省'"`
	City         string `gorm:"type:varchar(10) comment '市'"`
	District     string `gorm:"type:varchar(20) comment '区'"`
	Address      string `gorm:"type:varchar(100) comment '详细地址'"`
	SignerName   string `gorm:"type:varchar(20) comment '收货人姓名'"`
	SignerMobile string `gorm:"type:varchar(11) comment '收货人手机号码'"`
}

// 商品收藏表
type UserFav struct {
	BaseModel

	User  int32 `gorm:"index:idx_user_goods,unique;not null;type:int comment '用户ID'"`
	Goods int32 `gorm:"index:idx_user_goods,unique;not null;type:int comment '商品ID'"`
}
