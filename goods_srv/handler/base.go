package handler

import (
	"shop_srvs/goods_srv/proto"

	"gorm.io/gorm"
)

// 分页
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

type GoodsServer struct {
	proto.UnimplementedGoodsServer
}
