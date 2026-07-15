package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/emptypb"

	userpb "github.com/Jedeft/demo-micro-base-user/api/protobuf"

	"github.com/Jedeft/demo-micro-gw-admin/internal/mocks"
)

func TestInit_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := mocks.NewMockUserServiceClient(ctrl)
	orig := newUserClient
	t.Cleanup(func() { newUserClient = orig })
	newUserClient = func() (userpb.UserServiceClient, error) { return mock, nil }

	Init()
	assert.Same(t, mock, UserService.UserClient)
}

func TestUserSrv_Login_EmptyResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := mocks.NewMockUserServiceClient(ctrl)
	mock.EXPECT().Get(gomock.Any(), gomock.Any()).
		Return(&userpb.UserRow{}, nil)
	mock.EXPECT().Update(gomock.Any(), gomock.Any()).
		Return(&emptypb.Empty{}, nil)

	srv := UserSrv{UserClient: mock}
	out, err := srv.Login(context.Background(), "u", "p", "127.0.0.1")
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, uint32(0), out.Id)
}

func TestUserSrv_Login_ContextCancel(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := mocks.NewMockUserServiceClient(ctrl)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mock.EXPECT().Get(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("context canceled"))

	srv := UserSrv{UserClient: mock}
	out, err := srv.Login(ctx, "u", "p", "127.0.0.1")
	assert.Error(t, err)
	assert.Nil(t, out)
}
