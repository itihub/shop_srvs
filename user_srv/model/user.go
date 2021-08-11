package model

import (
	"gorm.io/gorm"
	"time"
)

// 数据表公共字段
type BaseModel struct {
	ID int32 `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"column:add_time"`	// 自定义字段名
	UpdatedAt time.Time `gorm:"column:update_time"`
	DeletedAt gorm.DeletedAt	// gorm 逻辑删除字段
	IsDeleted bool `gorm:"column:is_deleted"`
}

/*
1. 密码：密文且不可逆保存，防止丢失
	1. 对称加密 可逆
	2. 非对称加密 可逆
	3. md5 信息摘要算法 不可逆
	密码不可以反解，用户找回密码怎么办？ 只能修改密码
 */
type User struct {
	BaseModel
	Mobile string `gorm:"index:idx_mobile;unique;not null;type:varchar(11) comment '手机号码'"`
	Password string `gorm:"not null;type:varchar(100) comment '密码'"`
	NickName string `gorm:"type:varchar(20) comment '昵称'"`
	Birthday *time.Time `gorm:"type:datetime comment '生日'"`
	Gender string `gorm:"column:gender;default:male;type:varchar(6) comment 'female:女 male:男'"`
	Role int `gorm:"column:role;default:1;type:int comment '1:普通用户 2:管理员'"`
}
