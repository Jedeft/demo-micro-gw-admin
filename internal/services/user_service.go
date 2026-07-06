package services

import (
	"context"

	"github.com/Jedeft/xuanwu/log"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	userpb "github.com/Jedeft/demo-micro-base-user/api/protobuf"

	"github.com/Jedeft/demo-micro-gw-admin/internal/grpc"
)

// UserSrv 用户服务层
//
// handlers -> services -> 下游业务服务 gRPC 调用
type UserSrv struct {
	// UserClient 下游 user 服务的原生 gRPC client
	UserClient userpb.UserServiceClient
}

// UserService 全局用户服务层实例
var UserService UserSrv

// newUserClient 创建下游 user gRPC client；测试中可替换以隔离 gRPC 连接依赖。
var newUserClient = func() (userpb.UserClient, error) {
	conn, err := grpc.GetConn("demo-base-user")
	if err != nil {
		return nil, err
	}
	return userpb.NewUserClient(conn), nil
}

// InitUser 初始化用户服务层，创建下游 gRPC client
func InitUser() {
	client, err := newUserClient()
	if err != nil {
		log.Get().Bg().Fatal("init user service: get grpc conn failed", zap.Error(err))
	}
	UserService = UserSrv{
		UserClient: userpb.NewUserServiceClient(conn),
	}
}

// Login 用户登录：凭用户名+密码获取用户信息，并异步更新最后登录 IP
func (s *UserSrv) Login(ctx context.Context, username, password, loginIP string) (*userpb.UserRow, error) {
	out, err := s.UserClient.Get(ctx, &userpb.GetUserRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, err
	}

	// 登录成功后异步更新登录 IP（不影响主流程）
	_, updateErr := s.UserClient.Update(ctx, &userpb.UpdateUserRequest{
		Id:            out.Id,
		LastLoginIp:   loginIP,
		UpdatedUserId: out.Id,
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{"last_login_ip"},
		},
	})
	if updateErr != nil {
		log.Get().Bg().Error("user login update system error", zap.Error(updateErr))
	}

	return out, nil
}
