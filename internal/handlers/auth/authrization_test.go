package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	userpb "github.com/Jedeft/demo-micro-base-user/api/protobuf"
	"github.com/Jedeft/xuanwu/log"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/Jedeft/demo-micro-gw-admin/internal/configs"
	"github.com/Jedeft/demo-micro-gw-admin/internal/handlers"
	"github.com/Jedeft/demo-micro-gw-admin/internal/mocks"
	"github.com/Jedeft/demo-micro-gw-admin/internal/services"
	"github.com/Jedeft/demo-micro-gw-admin/internal/supports"
	"github.com/Jedeft/demo-micro-gw-admin/internal/testutil"
)

// newTestAuthHandler 构造一个带 mock UserClient 的 AuthrizationHandler。
func newTestAuthHandler(t *testing.T) (*AuthrizationHandler, *mocks.MockUserClient) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mock := mocks.NewMockUserClient(ctrl)
	h := &AuthrizationHandler{
		log:     log.Get(),
		userSrv: services.UserSrv{UserClient: mock},
	}
	return h, mock
}

func newAuthJSONContext(method, path, body string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath(path)
	return c, rec
}

func setUserContext(c echo.Context) {
	c.Set(handlers.ParsedUserKey, &handlers.UserInfo{
		ID:       1,
		Username: "admin",
		Name:     "Admin",
	})
}

func TestLoginReq_valid(t *testing.T) {
	testutil.SetTestConfig()
	configs.Config.Debug.IsFilterCaptcha = true

	tests := []struct {
		name     string
		req      LoginReq
		wantErr  bool
		wantCode int
	}{
		{name: "missing username", req: LoginReq{Password: "p"}, wantErr: true, wantCode: supports.UsernamePwdError},
		{name: "missing password", req: LoginReq{Username: "u"}, wantErr: true, wantCode: supports.UsernamePwdError},
		{name: "valid filter captcha", req: LoginReq{Username: "u", Password: "p"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.valid()
			if tt.wantErr {
				require.NotNil(t, err)
				assert.Equal(t, uint32(tt.wantCode), err.Code)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestLoginReq_valid_WithCaptcha(t *testing.T) {
	testutil.SetTestConfig()
	configs.Config.Debug.IsFilterCaptcha = false

	tests := []struct {
		name    string
		req     LoginReq
		wantErr bool
	}{
		{name: "missing captcha", req: LoginReq{Username: "u", Password: "p"}, wantErr: true},
		{name: "valid with captcha", req: LoginReq{Username: "u", Password: "p", Captcha: "abc"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.valid()
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestAuthrizationHandler_Login_Success(t *testing.T) {
	testutil.SetTestConfig()
	testutil.StartRedis(t)

	h, mock := newTestAuthHandler(t)
	body := `{"username":"admin","password":"123456","captcha":"abc"}`
	c, rec := newAuthJSONContext(http.MethodPost, "/v1/login", body)

	mock.EXPECT().Get(gomock.Any(), gomock.Any()).
		Return(&userpb.UserRow{ID: 1, Username: "admin", Name: "Admin"}, nil)
	mock.EXPECT().Update(gomock.Any(), gomock.Any()).
		Return(&emptypb.Empty{}, nil)

	err := h.Login(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var rsp handlers.RspRow
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &rsp))
	assert.NotNil(t, rsp.Row)
}

func TestAuthrizationHandler_Login_NotFound(t *testing.T) {
	testutil.SetTestConfig()
	testutil.StartRedis(t)

	h, mock := newTestAuthHandler(t)
	body := `{"username":"admin","password":"123456","captcha":"abc"}`
	c, _ := newAuthJSONContext(http.MethodPost, "/v1/login", body)

	mock.EXPECT().Get(gomock.Any(), gomock.Any()).
		Return(nil, status.Error(codes.NotFound, "not found"))

	err := h.Login(c)
	assert.Error(t, err)
}

func TestAuthrizationHandler_Login_PermissionDenied(t *testing.T) {
	testutil.SetTestConfig()
	testutil.StartRedis(t)

	h, mock := newTestAuthHandler(t)
	body := `{"username":"admin","password":"wrong","captcha":"abc"}`
	c, _ := newAuthJSONContext(http.MethodPost, "/v1/login", body)

	mock.EXPECT().Get(gomock.Any(), gomock.Any()).
		Return(nil, status.Error(codes.PermissionDenied, "denied"))

	err := h.Login(c)
	assert.Error(t, err)
}

func TestAuthrizationHandler_Login_OtherError(t *testing.T) {
	testutil.SetTestConfig()
	testutil.StartRedis(t)

	h, mock := newTestAuthHandler(t)
	body := `{"username":"admin","password":"123456","captcha":"abc"}`
	c, _ := newAuthJSONContext(http.MethodPost, "/v1/login", body)

	mock.EXPECT().Get(gomock.Any(), gomock.Any()).
		Return(nil, status.Error(codes.Internal, "rpc error"))

	err := h.Login(c)
	assert.Error(t, err)
}

func TestAuthrizationHandler_Login_BindError(t *testing.T) {
	testutil.SetTestConfig()
	testutil.StartRedis(t)

	h, _ := newTestAuthHandler(t)
	c, _ := newAuthJSONContext(http.MethodPost, "/v1/login", "invalid json")

	err := h.Login(c)
	assert.Error(t, err)
}

func TestAuthrizationHandler_Login_ValidationError(t *testing.T) {
	testutil.SetTestConfig()
	testutil.StartRedis(t)

	h, _ := newTestAuthHandler(t)
	body := `{"username":"","password":"","captcha":""}`
	c, _ := newAuthJSONContext(http.MethodPost, "/v1/login", body)

	err := h.Login(c)
	assert.Error(t, err)
}

func TestAuthrizationHandler_Logout_Success(t *testing.T) {
	testutil.SetTestConfig()
	testutil.StartRedis(t)

	h, _ := newTestAuthHandler(t)
	c, rec := newAuthJSONContext(http.MethodPost, "/v1/logout", "")
	setUserContext(c)

	err := h.Logout(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthrizationHandler_Logout_RedisError(t *testing.T) {
	testutil.SetTestConfig()

	h, _ := newTestAuthHandler(t)
	c, _ := newAuthJSONContext(http.MethodPost, "/v1/logout", "")
	setUserContext(c)

	err := h.Logout(c)
	assert.Error(t, err)
}
