package services

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/emptypb"

	userpb "github.com/Jedeft/demo-micro-base-user/api/protobuf"

	"github.com/Jedeft/demo-micro-gw-admin/internal/mocks"
)

// newMockClient 构造一个 gomock controller 与对应的 MockUserServiceClient，并在测试结束自动清理。
func newMockClient(t *testing.T) (*gomock.Controller, *mocks.MockUserServiceClient) {
	t.Helper()
	ctrl := gomock.NewController(t)
	return ctrl, mocks.NewMockUserServiceClient(ctrl)
}

func TestUserSrv_Login_GetError(t *testing.T) {
	_, mock := newMockClient(t)
	mock.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, errors.New("rpc get failed"))
	// Update 不应被调用：不设置任何 EXPECT，gomock 在未匹配调用时报错。

	srv := UserSrv{UserClient: mock}
	out, err := srv.Login(context.Background(), "u", "p", "127.0.0.1")
	require.Error(t, err)
	assert.Equal(t, "rpc get failed", err.Error())
	assert.Nil(t, out)
}

func TestUserSrv_Login_UpdateError(t *testing.T) {
	_, mock := newMockClient(t)
	mock.EXPECT().Get(gomock.Any(), gomock.Any()).
		Return(&userpb.UserRow{Id: 1, Username: "u"}, nil)
	// Update 失败仅记录日志，不影响主流程返回 out。
	mock.EXPECT().Update(gomock.Any(), gomock.Any()).
		Return(&emptypb.Empty{}, errors.New("rpc update failed"))

	srv := UserSrv{UserClient: mock}
	out, err := srv.Login(context.Background(), "u", "p", "127.0.0.1")
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, uint32(1), out.Id)
	assert.Equal(t, "u", out.Username)
}

func TestUserSrv_Login_Success(t *testing.T) {
	_, mock := newMockClient(t)
	mock.EXPECT().Get(gomock.Any(), gomock.Any()).
		Return(&userpb.UserRow{Id: 2, Username: "u", Name: "n"}, nil)
	mock.EXPECT().Update(gomock.Any(), gomock.Any()).Return(&emptypb.Empty{}, nil)

	srv := UserSrv{UserClient: mock}
	out, err := srv.Login(context.Background(), "u", "p", "10.0.0.1")
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, uint32(2), out.Id)
	assert.Equal(t, "n", out.Name)
}

func TestInitUser_Success(t *testing.T) {
	_, mock := newMockClient(t)
	orig := newUserClient
	t.Cleanup(func() { newUserClient = orig })
	newUserClient = func() (userpb.UserServiceClient, error) { return mock, nil }

	InitUser()
	assert.Same(t, mock, UserService.UserClient)
}

// TestInitUser_Fatal 通过子进程验证：newUserClient 返回 err 时 InitUser 会 Fatal 退出（非 0）。
func TestInitUser_Fatal(t *testing.T) {
	if os.Getenv("TEST_INITUSER_FATAL") == "1" {
		orig := newUserClient
		newUserClient = func() (userpb.UserServiceClient, error) { return nil, errors.New("dial failed") }
		defer func() { newUserClient = orig }()
		InitUser() // 期望在此 Fatal 退出
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestInitUser_Fatal")
	cmd.Env = append(os.Environ(), "TEST_INITUSER_FATAL=1")
	err := cmd.Run()
	require.Error(t, err, "子进程应以非 0 退出码结束")
}
