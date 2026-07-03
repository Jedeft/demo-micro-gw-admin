package users

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
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/Jedeft/demo-micro-gw-admin/internal/handlers"
	"github.com/Jedeft/demo-micro-gw-admin/internal/mocks"
	"github.com/Jedeft/demo-micro-gw-admin/internal/services"
)

// newTestUserHandler 构造一个带 mock UserClient 的 UserHandler。
func newTestUserHandler(t *testing.T) (*UserHandler, *mocks.MockUserClient) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mock := mocks.NewMockUserClient(ctrl)
	h := &UserHandler{
		log:     log.Get(),
		userSrv: services.UserSrv{UserClient: mock},
	}
	return h, mock
}

// newJSONContext 构造一个携带 JSON body 的 echo.Context，返回 context 与 recorder。
func newJSONContext(method, path, body string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath(path)
	return c, rec
}

// newQueryContext 构造一个携带 query 参数的 echo.Context，返回 context 与 recorder。
func newQueryContext(method, target, path string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, target, nil)
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath(path)
	return c, rec
}

// setUserContext 在 echo.Context 中注入已登录用户信息。
func setUserContext(c echo.Context) {
	c.Set(handlers.ParsedUserKey, &handlers.UserInfo{
		ID:       1,
		Username: "admin",
		Name:     "Admin",
	})
}

func TestAddUserReq_valid(t *testing.T) {
	tests := []struct {
		name    string
		req     AddUserReq
		wantErr bool
	}{
		{name: "missing username", req: AddUserReq{Name: "n", Phone: "p"}, wantErr: true},
		{name: "missing password", req: AddUserReq{Username: "u", Name: "n", Phone: "p"}, wantErr: true},
		{name: "missing name", req: AddUserReq{Username: "u", Password: "p", Phone: "ph"}, wantErr: true},
		{name: "missing phone", req: AddUserReq{Username: "u", Password: "p", Name: "n"}, wantErr: true},
		{name: "valid", req: AddUserReq{Username: "u", Password: "p", Name: "n", Phone: "ph"}, wantErr: false},
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

func TestUpdateUserReq_valid(t *testing.T) {
	tests := []struct {
		name    string
		req     UpdateUserReq
		wantErr bool
	}{
		{name: "zero id", req: UpdateUserReq{Name: "n", Phone: "p"}, wantErr: true},
		{name: "missing name", req: UpdateUserReq{ID: 1, Phone: "p"}, wantErr: true},
		{name: "missing phone", req: UpdateUserReq{ID: 1, Name: "n"}, wantErr: true},
		{name: "valid", req: UpdateUserReq{ID: 1, Name: "n", Phone: "p"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.valid()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestChangePWDReq_valid(t *testing.T) {
	tests := []struct {
		name    string
		req     ChangePWDReq
		wantErr bool
	}{
		{name: "zero id", req: ChangePWDReq{OldPassword: "a", NewPassword: "b"}, wantErr: true},
		{name: "same password", req: ChangePWDReq{ID: 1, OldPassword: "a", NewPassword: "a"}, wantErr: true},
		{name: "valid", req: ChangePWDReq{ID: 1, OldPassword: "a", NewPassword: "b"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.valid()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserHandler_Add_Success(t *testing.T) {
	h, mock := newTestUserHandler(t)
	body := `{"username":"u","password":"p","name":"n","phone":"133"}`
	c, rec := newJSONContext(http.MethodPost, "/v1/user/add", body)
	setUserContext(c)

	mock.EXPECT().Create(gomock.Any(), gomock.Any()).
		Return(&userpb.CreateUseResp{ID: 1}, nil)

	err := h.Add(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUserHandler_Add_BindError(t *testing.T) {
	h, _ := newTestUserHandler(t)
	c, _ := newJSONContext(http.MethodPost, "/v1/user/add", "invalid json")

	err := h.Add(c)
	assert.Error(t, err)
}

func TestUserHandler_Add_ValidationError(t *testing.T) {
	h, _ := newTestUserHandler(t)
	body := `{"username":"","password":"","name":"","phone":""}`
	c, _ := newJSONContext(http.MethodPost, "/v1/user/add", body)
	setUserContext(c)

	err := h.Add(c)
	assert.Error(t, err)
}

func TestUserHandler_Add_RPCError(t *testing.T) {
	h, mock := newTestUserHandler(t)
	body := `{"username":"u","password":"p","name":"n","phone":"133"}`
	c, _ := newJSONContext(http.MethodPost, "/v1/user/add", body)
	setUserContext(c)

	mock.EXPECT().Create(gomock.Any(), gomock.Any()).
		Return(nil, status.Error(codes.Internal, "rpc error"))

	err := h.Add(c)
	assert.Error(t, err)
}

func TestUserHandler_Get_Success(t *testing.T) {
	h, mock := newTestUserHandler(t)
	c, rec := newQueryContext(http.MethodGet, "/v1/user?id=1", "/v1/user")

	mock.EXPECT().Get(gomock.Any(), gomock.Any()).
		Return(&userpb.UserRow{ID: 1, Username: "u", Name: "n", Phone: "133"}, nil)

	err := h.Get(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var rsp handlers.RspRow
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &rsp))
	assert.Equal(t, uint32(0), rsp.Code)
}

func TestUserHandler_Get_NotFound(t *testing.T) {
	h, mock := newTestUserHandler(t)
	c, _ := newQueryContext(http.MethodGet, "/v1/user?id=999", "/v1/user")

	mock.EXPECT().Get(gomock.Any(), gomock.Any()).
		Return(nil, status.Error(codes.NotFound, "not found"))

	err := h.Get(c)
	assert.Error(t, err)
}

func TestUserHandler_Get_ValidationError(t *testing.T) {
	h, _ := newTestUserHandler(t)
	c, _ := newQueryContext(http.MethodGet, "/v1/user?id=0", "/v1/user")

	err := h.Get(c)
	assert.Error(t, err)
}

func TestUserHandler_Get_RPCError(t *testing.T) {
	h, mock := newTestUserHandler(t)
	c, _ := newQueryContext(http.MethodGet, "/v1/user?id=1", "/v1/user")

	mock.EXPECT().Get(gomock.Any(), gomock.Any()).
		Return(nil, status.Error(codes.Internal, "rpc error"))

	err := h.Get(c)
	assert.Error(t, err)
}

func TestUserHandler_List_Success(t *testing.T) {
	h, mock := newTestUserHandler(t)
	c, rec := newQueryContext(http.MethodGet, "/v1/user/list?limit=10&offset=0", "/v1/user/list")

	rows := []*userpb.UserRow{
		{ID: 1, Username: "u1", Name: "n1", Phone: "111", CreatedUserID: 10},
		{ID: 2, Username: "u2", Name: "n2", Phone: "222", CreatedUserID: 10},
	}
	mock.EXPECT().List(gomock.Any(), gomock.Any()).
		Return(&userpb.ListUserResp{Rows: rows, Total: 2}, nil)
	mock.EXPECT().Search(gomock.Any(), gomock.Any()).
		Return(&userpb.SearchUserResp{
			Rows: []*userpb.UserRow{{ID: 10, Name: "creator"}},
		}, nil)

	err := h.List(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var rsp handlers.RspRows
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &rsp))
	assert.Equal(t, int64(2), rsp.Total)
}

func TestUserHandler_List_DefaultLimit(t *testing.T) {
	h, mock := newTestUserHandler(t)
	c, _ := newQueryContext(http.MethodGet, "/v1/user/list?limit=0&offset=0", "/v1/user/list")

	mock.EXPECT().List(gomock.Any(), gomock.Any()).
		Return(&userpb.ListUserResp{Rows: nil, Total: 0}, nil)
	mock.EXPECT().Search(gomock.Any(), gomock.Any()).
		Return(&userpb.SearchUserResp{}, nil)

	err := h.List(c)
	require.NoError(t, err)
}

func TestUserHandler_List_MaxLimitClamp(t *testing.T) {
	h, mock := newTestUserHandler(t)
	c, _ := newQueryContext(http.MethodGet, "/v1/user/list?limit=200&offset=0", "/v1/user/list")

	mock.EXPECT().List(gomock.Any(), gomock.Any()).
		Return(&userpb.ListUserResp{Rows: nil, Total: 0}, nil)
	mock.EXPECT().Search(gomock.Any(), gomock.Any()).
		Return(&userpb.SearchUserResp{}, nil)

	err := h.List(c)
	require.NoError(t, err)
}

func TestUserHandler_List_RPCError(t *testing.T) {
	h, mock := newTestUserHandler(t)
	c, _ := newQueryContext(http.MethodGet, "/v1/user/list", "/v1/user/list")

	mock.EXPECT().List(gomock.Any(), gomock.Any()).
		Return(nil, status.Error(codes.Internal, "rpc error"))

	err := h.List(c)
	assert.Error(t, err)
}

func TestUserHandler_Search_Success(t *testing.T) {
	h, mock := newTestUserHandler(t)
	c, rec := newQueryContext(http.MethodGet, "/v1/user/search?name=a&phone=1", "/v1/user/search")

	mock.EXPECT().Search(gomock.Any(), gomock.Any()).
		Return(&userpb.SearchUserResp{
			Rows:  []*userpb.UserRow{{ID: 1, Username: "u", Name: "n", Phone: "133"}},
			Total: 1,
		}, nil)

	err := h.Search(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var rsp handlers.RspRows
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &rsp))
	assert.Equal(t, int64(1), rsp.Total)
}

func TestUserHandler_Search_RPCError(t *testing.T) {
	h, mock := newTestUserHandler(t)
	c, _ := newQueryContext(http.MethodGet, "/v1/user/search", "/v1/user/search")

	mock.EXPECT().Search(gomock.Any(), gomock.Any()).
		Return(nil, status.Error(codes.Internal, "rpc error"))

	err := h.Search(c)
	assert.Error(t, err)
}

func TestUserHandler_Update_Success(t *testing.T) {
	h, mock := newTestUserHandler(t)
	body := `{"id":1,"name":"n","phone":"133","partner":"p","note":"note"}`
	c, rec := newJSONContext(http.MethodPost, "/v1/user/update", body)
	setUserContext(c)

	mock.EXPECT().Update(gomock.Any(), gomock.Any()).
		Return(&emptypb.Empty{}, nil)

	err := h.Update(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUserHandler_Update_NotFound(t *testing.T) {
	h, mock := newTestUserHandler(t)
	body := `{"id":999,"name":"n","phone":"133"}`
	c, _ := newJSONContext(http.MethodPost, "/v1/user/update", body)
	setUserContext(c)

	mock.EXPECT().Update(gomock.Any(), gomock.Any()).
		Return(nil, status.Error(codes.NotFound, "not found"))

	err := h.Update(c)
	assert.Error(t, err)
}

func TestUserHandler_Update_ValidationError(t *testing.T) {
	h, _ := newTestUserHandler(t)
	body := `{"id":0,"name":"","phone":""}`
	c, _ := newJSONContext(http.MethodPost, "/v1/user/update", body)
	setUserContext(c)

	err := h.Update(c)
	assert.Error(t, err)
}

func TestUserHandler_Update_RPCError(t *testing.T) {
	h, mock := newTestUserHandler(t)
	body := `{"id":1,"name":"n","phone":"133"}`
	c, _ := newJSONContext(http.MethodPost, "/v1/user/update", body)
	setUserContext(c)

	mock.EXPECT().Update(gomock.Any(), gomock.Any()).
		Return(nil, status.Error(codes.Internal, "rpc error"))

	err := h.Update(c)
	assert.Error(t, err)
}

func TestUserHandler_ChangePWD_Success(t *testing.T) {
	h, mock := newTestUserHandler(t)
	body := `{"id":1,"old_password":"a","new_password":"b"}`
	c, rec := newJSONContext(http.MethodPost, "/v1/user/password/update", body)
	setUserContext(c)

	mock.EXPECT().ChangePWD(gomock.Any(), gomock.Any()).
		Return(&emptypb.Empty{}, nil)

	err := h.ChangePWD(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUserHandler_ChangePWD_ValidationError(t *testing.T) {
	h, _ := newTestUserHandler(t)
	body := `{"id":1,"old_password":"a","new_password":"a"}`
	c, _ := newJSONContext(http.MethodPost, "/v1/user/password/update", body)
	setUserContext(c)

	err := h.ChangePWD(c)
	assert.Error(t, err)
}

func TestUserHandler_ChangePWD_RPCError(t *testing.T) {
	h, mock := newTestUserHandler(t)
	body := `{"id":1,"old_password":"a","new_password":"b"}`
	c, _ := newJSONContext(http.MethodPost, "/v1/user/password/update", body)
	setUserContext(c)

	mock.EXPECT().ChangePWD(gomock.Any(), gomock.Any()).
		Return(nil, status.Error(codes.Internal, "rpc error"))

	err := h.ChangePWD(c)
	assert.Error(t, err)
}

func TestUserHandler_Delete_Success(t *testing.T) {
	h, mock := newTestUserHandler(t)
	c, rec := newQueryContext(http.MethodGet, "/v1/user/delete?id=2", "/v1/user/delete")
	setUserContext(c)

	mock.EXPECT().Delete(gomock.Any(), gomock.Any()).
		Return(&emptypb.Empty{}, nil)

	err := h.Delete(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUserHandler_Delete_ValidationError(t *testing.T) {
	h, _ := newTestUserHandler(t)
	c, _ := newQueryContext(http.MethodGet, "/v1/user/delete?id=0", "/v1/user/delete")
	setUserContext(c)

	err := h.Delete(c)
	assert.Error(t, err)
}

func TestUserHandler_Delete_RPCError(t *testing.T) {
	h, mock := newTestUserHandler(t)
	c, _ := newQueryContext(http.MethodGet, "/v1/user/delete?id=2", "/v1/user/delete")
	setUserContext(c)

	mock.EXPECT().Delete(gomock.Any(), gomock.Any()).
		Return(nil, status.Error(codes.Internal, "rpc error"))

	err := h.Delete(c)
	assert.Error(t, err)
}

func TestAssUserList_Empty(t *testing.T) {
	h, mock := newTestUserHandler(t)
	c, _ := newQueryContext(http.MethodGet, "/", "/")

	mock.EXPECT().Search(gomock.Any(), gomock.Any()).
		Return(&userpb.SearchUserResp{}, nil)

	rows, err := h.assUserList(nil, c)
	require.NoError(t, err)
	assert.Empty(t, rows)
}

func TestUpdateUserReq_BuildPartnerNote(t *testing.T) {
	req := UpdateUserReq{ID: 1, Name: "n", Phone: "p", Partner: "partner", Note: "note"}
	var partner, note *wrapperspb.StringValue
	if req.Partner != "" {
		partner = wrapperspb.String(req.Partner)
	}
	if req.Note != "" {
		note = wrapperspb.String(req.Note)
	}
	assert.Equal(t, "partner", partner.GetValue())
	assert.Equal(t, "note", note.GetValue())
}
