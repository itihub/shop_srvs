package model

// 商品分类表
type Category struct {
	BaseModel
	Name             string `gorm:"not null;type:varchar(20) comment '名称'"`
	ParentCategoryID int32
	ParentCategory   *Category
	Level            int32 `gorm:"default:1;not null;type:int comment '级别'"`
	IsTab            bool  `gorm:"not null;default:false;type:bool  comment '是否展示在Tab栏'"`
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
