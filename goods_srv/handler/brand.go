package handler

import (
	"context"
	"shop_srvs/goods_srv/global"
	"shop_srvs/goods_srv/model"
	"shop_srvs/goods_srv/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *GoodsServer) BrandList(ctx context.Context, req *proto.BrandFilterRequest) (*proto.BrandListResponse, error) {
	// 获取品牌列表
	brandListResponse := proto.BrandListResponse{}

	var brands []model.Brand
	result := global.DB.Scopes(Paginate(int(req.Pages), int(req.PagePerNums))).Find(&brands)
	if result.Error != nil {
		return nil, result.Error
	}

	var count int64
	global.DB.Model(&model.Brand{}).Count(&count)
	brandListResponse.Total = int32(count)

	var brandResponses []*proto.BrandInfoResponse
	for _, brand := range brands {
		brandResponses = append(brandResponses, &proto.BrandInfoResponse{
			Id:   brand.ID,
			Name: brand.Name,
			Logo: brand.Logo,
		})
	}

	brandListResponse.Data = brandResponses

	return &brandListResponse, nil

}
func (s *GoodsServer) CreateBrand(ctx context.Context, req *proto.BrandRequest) (*proto.BrandInfoResponse, error) {
	// 新建品牌
	if result := global.DB.Where("name=?", req.Name).First(&model.Brand{}); result.RowsAffected > 0 {
		return nil, status.Errorf(codes.InvalidArgument, "品牌已存在")
	}

	brand := &model.Brand{
		Name: req.Name,
		Logo: req.Logo,
	}

	global.DB.Save(brand)
	return &proto.BrandInfoResponse{
		Id: brand.ID,
	}, nil
}

func (s *GoodsServer) DeleteBrand(ctx context.Context, req *proto.BrandRequest) (*emptypb.Empty, error) {
	// 删除品牌
	if result := global.DB.Delete(&model.Brand{}, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "品牌不存在")
	}
	return &emptypb.Empty{}, nil
}

func (s *GoodsServer) UpdateBrand(ctx context.Context, req *proto.BrandRequest) (*emptypb.Empty, error) {
	// 更新品牌
	brand := model.Brand{}
	if result := global.DB.First(&brand, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "品牌不存在")
	}

	if req.Name != "" {
		brand.Name = req.Name
	}
	if req.Logo != "" {
		brand.Logo = req.Logo
	}

	global.DB.Save(&brand)
	return &emptypb.Empty{}, nil
}
