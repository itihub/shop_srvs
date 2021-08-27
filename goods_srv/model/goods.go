package model

import (
	"context"
	"shop_srvs/goods_srv/global"
	"strconv"

	"gorm.io/gorm"
)

/*
	类型， 这个字段是否能为null， 这个字段应该设置为可以为null还是设置为空， 0
	实际开发过程中 尽量设置为不为null
	https://zhuanlan.zhihu.com/p/73997266
*/

// 商品分类表
type Category struct {
	BaseModel
	Name             string      `gorm:"not null;type:varchar(20) comment '名称'" json:"name"`
	ParentCategoryID int32       `json:"parent"`
	ParentCategory   *Category   `json:"-"` // - 序列化会忽略
	SubCategory      []*Category `gorm:"foreignKey:ParentCategoryID;references:ID" json:"sub_category"`
	Level            int32       `gorm:"default:1;not null;type:int comment '级别'" json:"level"`
	IsTab            bool        `gorm:"not null;default:false;type:bool  comment '是否展示在Tab栏'" json:"is_tab"`
}

// 品牌表
type Brand struct {
	BaseModel
	Name string `gorm:"not null;type:varchar(20) comment '名称'"`
	Logo string `gorm:"not null;default:'';type:varchar(200) comment '图标'"`
}

// 商品分类与品牌关联表
type GoodsCategoryBrand struct {
	BaseModel

	CategoryID int32 `gorm:"type:int;index:idx_category_brand,unique"`
	Category   Category

	BrandID int32 `gorm:"type:int;index:idx_category_brand,unique"`
	Brand   Brand
}

// 轮播图表
type Banner struct {
	BaseModel
	Image string `gorm:"not null;type:varchar(200) comment '图片地址'"`
	Url   string `gorm:"not null;type:varchar(200) comment '图片地址'"`
	Index int32  `gorm:"default:1;not null;type:int comment '序号'"`
}

// 商品表
type Goods struct {
	BaseModel

	CategoryID int32 `gorm:"type:int;not null;"`
	Category   Category
	BrandID    int32 `gorm:"type:int;not null;"`
	Brand      Brand

	OnSale   bool `gorm:"default:false;not null;type:bool comment '是否上架'"`
	ShipFree bool `gorm:"default:false;not null;type:bool comment '是否免邮费'"`
	IsNew    bool `gorm:"default:false;not null;type:bool comment '是否新品'"`
	IsHot    bool `gorm:"default:false;not null;type:bool comment '是否热卖商品'"`

	Name            string   `gorm:"not null;type:varchar(50) comment '名称'"`
	GoodsSn         string   `gorm:"not null;type:varchar(50) comment '编号'"`
	ClickNum        int32    `gorm:"default:0;not null;type:int comment '点击数'"`
	SoldNum         int32    `gorm:"default:0;not null;type:int comment '销量'"`
	FavNum          int32    `gorm:"default:0;not null;type:int comment '收藏数'"`
	MarketPrice     float32  `gorm:"not null;type:float comment '指导价'"`
	ShopPrice       float32  `gorm:"not null;type:float comment '售价'"`
	GoodsBrief      string   `gorm:"not null;type:varchar(100) comment '商品简介'"`
	Images          GormList `gorm:"not null;type:varchar(1000) comment '商品图片'"`
	DescImages      GormList `gorm:"not null;type:varchar(1000) comment '商品详情图片'"`
	GoodsFrontImage string   `gorm:"not null;type:varchar(200) comment '商品主图地址'"`
}

// gorm 钩子方法 降低代码耦合性 商品创建成功后同步到es中
func (g *Goods) AfterCreate(tx *gorm.DB) (err error) {
	esModel := EsGoods{
		ID:          g.ID,
		CategoryID:  g.CategoryID,
		BrandID:     g.BrandID,
		OnSale:      g.OnSale,
		ShipFree:    g.ShipFree,
		IsNew:       g.IsNew,
		IsHot:       g.IsHot,
		Name:        g.Name,
		ClickNum:    g.ClickNum,
		SoldNum:     g.SoldNum,
		FavNum:      g.FavNum,
		MarketPrice: g.MarketPrice,
		GoodsBrief:  g.GoodsBrief,
		ShopPrice:   g.ShopPrice,
	}
	_, err = global.ElasticClient.Index().Index(esModel.GetIndexName()).BodyJson(esModel).Id(strconv.Itoa(int(g.ID))).Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (g *Goods) AfterUpdate(tx *gorm.DB) (err error) {
	esModel := EsGoods{
		ID:          g.ID,
		CategoryID:  g.CategoryID,
		BrandID:     g.BrandID,
		OnSale:      g.OnSale,
		ShipFree:    g.ShipFree,
		IsNew:       g.IsNew,
		IsHot:       g.IsHot,
		Name:        g.Name,
		ClickNum:    g.ClickNum,
		SoldNum:     g.SoldNum,
		FavNum:      g.FavNum,
		MarketPrice: g.MarketPrice,
		GoodsBrief:  g.GoodsBrief,
		ShopPrice:   g.ShopPrice,
	}

	_, err = global.ElasticClient.Update().Index(esModel.GetIndexName()).
		Doc(esModel).Id(strconv.Itoa(int(g.ID))).Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (g *Goods) AfterDelete(tx *gorm.DB) (err error) {
	_, err = global.ElasticClient.Delete().Index(EsGoods{}.GetIndexName()).Id(strconv.Itoa(int(g.ID))).Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}
