package handel

import (
	"context"
	"go_micro/user_srv/global"
	"go_micro/user_srv/model"
	"go_micro/user_srv/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

type UserService struct {
	
}

// 分页
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func (db *gorm.DB) *gorm.DB {
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

func Model2Response(user model.User) proto.UserInfoResponse {
	// 在GRPC的message中字段有默认值，不能赋值nil进去容易错误
	userInfoRsp := proto.UserInfoResponse{
		Id: user.ID,
		Password: user.Password,
		NickName: user.NickName,
		Gender: user.Gender,
		Role: int32(user.Role),
	}
	if user.Birthday != nil {
		userInfoRsp.Birthday = uint64(user.Birthday.Unix())
	}
	return userInfoRsp
}

func (s *UserService) GetUserList(ctx context.Context, req *proto.PageInfo) (*proto.UserListResponse, error) {
	// 获取用户列表
	var users []model.User
	result := global.DB.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}

	rsp := &proto.UserListResponse{}
	rsp.Total = int32(result.RowsAffected)

	// 分页查找
	global.DB.Scopes(Paginate(int(req.Pn), int(req.PSize))).Find(&users)

	// 模型转换
	for _, user := range users {
		userInfoRsp := Model2Response(user)
		rsp.Data = append(rsp.Data, &userInfoRsp)
	}

	return rsp, nil
}

func (s *UserService) GetUserByMobile(ctx context.Context, req *proto.MobileRequest) (*proto.UserInfoResponse, error) {
	// 通过mobile查询用户
	var user model.User
	result := global.DB.Where(&model.User{Mobile: req.Mobile}).First(&user)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "用户不存在")
	}
	if result.Error != nil {
		return nil, result.Error
	}
	userInfoRsp := Model2Response(user)
	return &userInfoRsp, nil
}

func (s *UserService) GetUserById(ctx context.Context, req *proto.IdRequest) (*proto.UserInfoResponse, error) {
	// 通过id查询用户
	var user model.User
	result := global.DB.First(&user, req.Id)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "用户不存在")
	}
	if result.Error != nil {
		return nil, result.Error
	}
	userInfoRsp := Model2Response(user)
	return &userInfoRsp, nil
}

func (s *UserService) CreateUser(ctx context.Context, req *proto.CreateUserInfo) (*proto.UserInfoResponse, error) {
	// 创建用户
	return nil, nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *proto.UpdateUserInfo) (*emptypb.Empty, error) {
	return nil, nil
}

func (s *UserService) CheckPassword(context.Context, *proto.PasswordCheckInfo) (*proto.CheckResponse, error) {
	return nil, nil
}