package users

import (
	"net/http"

	"github.com/Jedeft/xuanwu/log"
	"github.com/Jedeft/xuanwu/xerr"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"

	userpb "github.com/Jedeft/demo-micro-base-user/api/protobuf"

	"github.com/Jedeft/demo-micro-gw-admin/internal/handlers"
	"github.com/Jedeft/demo-micro-gw-admin/internal/services"
	"github.com/Jedeft/demo-micro-gw-admin/internal/supports"
	"github.com/Jedeft/demo-micro-gw-admin/internal/supports/utils"
)

const (
	// defaultUserListLimit 用户列表默认每页条数
	defaultUserListLimit = 20
	// maxUserListLimit 用户列表每页上限，防止资源耗尽与 uint32 溢出
	maxUserListLimit = 100
)

// UserHandler 用户handler
type UserHandler struct {
	log     log.Factory
	userSrv services.UserSrv
}

// NewUserHandler new UserHandler
func NewUserHandler() *UserHandler {
	return &UserHandler{
		log:     log.Get(),
		userSrv: services.UserService,
	}
}

// AddUserReq 用户创建结构
type AddUserReq struct {
	Username string `json:"username" validate:"required" example:"admin"`    // 账号
	Password string `json:"password" validate:"required" example:"123456"`   // 密码，前端请做SHA256转化，不传明文
	Name     string `json:"name" validate:"required" example:"张三"`           // 用户名称
	Phone    string `json:"phone" validate:"required" example:"13333333333"` // 手机号
	Partner  string `json:"partner" example:"李四"`                            // 合作伙伴
	Note     string `json:"note" example:"备注信息"`                             // 备注信息
}

func (p *AddUserReq) valid() *xerr.Err {
	if len(p.Username) == 0 {
		return xerr.Get().NewErr(supports.GetErrMsg(supports.MissingUsernameError))
	}
	if len(p.Password) == 0 {
		return xerr.Get().NewErr(supports.GetErrMsg(supports.MissingPwdError))
	}
	if len(p.Name) == 0 {
		return xerr.Get().NewErr(supports.GetErrMsg(supports.MissingNameError))
	}
	if len(p.Phone) == 0 {
		return xerr.Get().NewErr(supports.GetErrMsg(supports.MissingPhoneError))
	}
	return nil
}

// Add 创建用户
func (h *UserHandler) Add(c echo.Context) error {
	var (
		req AddUserReq
		rsp handlers.RspRow
	)
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := req.valid(); err != nil {
		return err
	}
	_, err := h.userSrv.UserClient.Create(c.Request().Context(), &userpb.CreateUserReq{
		Username:      req.Username,
		Password:      req.Password,
		Name:          req.Name,
		Phone:         req.Phone,
		CreatedUserID: handlers.GetUserInfo(c).ID,
	})
	if err != nil {
		h.log.Bg().Error("user create error", zap.Error(err))
		return err
	}
	return c.JSON(http.StatusOK, rsp)
}

// GetUserRsp 用户获取结构
type GetUserRsp struct {
	ID       uint32 `json:"id,omitempty"`
	Username string `json:"username,omitempty"`
	Name     string `json:"name,omitempty"`
	Phone    string `json:"phone,omitempty"`
}

// Get 获取用户
func (h *UserHandler) Get(c echo.Context) error {
	var (
		req handlers.Uint32IDReq
		rsp handlers.RspRow
	)
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := req.Valid(); err != nil {
		return err
	}
	out, err := h.userSrv.UserClient.Get(c.Request().Context(), &userpb.GetUserReq{
		ID: req.ID,
	})
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.NotFound {
			return xerr.Get().NewErr(supports.GetErrMsg(supports.MissingPKError))
		}
		return err
	}
	rsp.Row = GetUserRsp{
		ID:       out.ID,
		Username: out.Username,
		Name:     out.Name,
		Phone:    out.Phone,
	}
	return c.JSON(http.StatusOK, rsp)
}

// ListUserReq 用户列表检索
type ListUserReq struct {
	BeginCreatedAt int64  `query:"begin_created_at"`
	EndCreatedAt   int64  `query:"end_created_at"`
	Name           string `query:"name"`  // 模糊匹配
	Phone          string `query:"phone"` // 模糊匹配
	Limit          int    `query:"limit"`
	Offset         int    `query:"offset"`
}

// ListUserRsp 用户列表数据
type ListUserRsp struct {
	ID              uint32 `json:"id"`
	Name            string `json:"name"`
	Username        string `json:"username"`
	Phone           string `json:"phone"`
	CreatedUserID   uint32 `json:"created_user_id"`
	CreatedUserName string `json:"created_user_name"`
}

// List 列表获取用户
func (h *UserHandler) List(c echo.Context) error {
	var (
		req ListUserReq
		rsp handlers.RspRows
	)
	if err := c.Bind(&req); err != nil {
		return err
	}

	if req.Limit <= 0 {
		req.Limit = defaultUserListLimit
	}
	if req.Limit > maxUserListLimit {
		req.Limit = maxUserListLimit
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	out, err := h.userSrv.UserClient.List(c.Request().Context(), &userpb.ListUserReq{
		PageLimit:  uint32(req.Limit),  //nolint:gosec // G115: 已 clamp 至 [1,100]
		PageOffset: uint32(req.Offset), //nolint:gosec // G115: 已 floor 至 >=0
		Condition: &userpb.SearchUserReq{
			BeginCreatedAt: req.BeginCreatedAt,
			EndCreatedAt:   req.EndCreatedAt,
		},
	})
	if err != nil {
		return err
	}

	rows, err := h.assUserList(out.Rows, c)
	if err != nil {
		return err
	}
	rsp.Rows = rows
	rsp.Total = out.Total
	return c.JSON(http.StatusOK, rsp)
}

// assUserList 组装List数据
func (h *UserHandler) assUserList(list []*userpb.UserRow, c echo.Context) ([]ListUserRsp, error) {
	userIDs := make([]uint32, 0, len(list))
	for _, v := range list {
		userIDs = append(userIDs, v.CreatedUserID)
	}

	// 通过ID集合检索数据
	userNameOut, err := h.userSrv.UserClient.Search(c.Request().Context(), &userpb.SearchUserReq{
		IDs: utils.RemoveUint32Duplication(userIDs),
	})
	if err != nil {
		return nil, err
	}

	// 转化成Map字典，便于检索
	userNameDic := make(map[uint32]string)
	for _, v := range userNameOut.Rows {
		userNameDic[v.ID] = v.Name
	}
	rows := make([]ListUserRsp, 0, len(list))
	for _, v := range list {
		rows = append(rows, ListUserRsp{
			ID:              v.ID,
			Name:            v.Name,
			Username:        v.Username,
			Phone:           v.Phone,
			CreatedUserID:   v.CreatedUserID,
			CreatedUserName: userNameDic[v.CreatedUserID],
		})
	}
	return rows, nil
}

// SearchUserReq 用户下拉框检索
type SearchUserReq struct {
	Name  string `query:"name"`  // 模糊匹配
	Phone string `query:"phone"` // 模糊匹配
	Limit int    `query:"limit"`
}

// SearchUserRsp 用户下拉框数据
type SearchUserRsp struct {
	ID       uint32 `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Phone    string `json:"phone"`
}

// Search 检索用户
func (h *UserHandler) Search(c echo.Context) error {
	var (
		req SearchUserReq
		rsp handlers.RspRows
	)
	if err := c.Bind(&req); err != nil {
		return err
	}

	out, err := h.userSrv.UserClient.Search(c.Request().Context(), &userpb.SearchUserReq{
		UsernameLike: req.Name,
		PhoneLike:    req.Phone,
	})
	if err != nil {
		return err
	}
	rows := make([]SearchUserRsp, 0, len(out.Rows))
	for _, v := range out.Rows {
		rows = append(rows, SearchUserRsp{
			ID:       v.ID,
			Name:     v.Name,
			Username: v.Username,
			Phone:    v.Phone,
		})
	}
	rsp.Rows = rows
	rsp.Total = out.Total
	return c.JSON(http.StatusOK, rsp)
}

// UpdateUserReq 用户更新
type UpdateUserReq struct {
	ID      uint32 `json:"id"  validate:"required" example:"1"`
	Name    string `json:"name" validate:"required" example:"张三"`
	Phone   string `json:"phone" validate:"required" example:"13333333333"`
	Partner string `json:"partner" swaggertype:"string" example:"合伙人"`
	Note    string `json:"note" swaggertype:"string" example:"备注"`
}

func (p *UpdateUserReq) valid() error {
	if p.ID == 0 {
		return xerr.Get().NewErr(supports.GetErrMsg(supports.MissingPKError))
	}
	if len(p.Name) == 0 {
		return xerr.Get().NewErr(supports.GetErrMsg(supports.MissingNameError))
	}
	if len(p.Phone) == 0 {
		return xerr.Get().NewErr(supports.GetErrMsg(supports.MissingPhoneError))
	}
	return nil
}

// Update 用户更新
func (h *UserHandler) Update(c echo.Context) error {
	var (
		req UpdateUserReq
		rsp handlers.RspRow
	)
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := req.valid(); err != nil {
		return err
	}

	var partner, note *wrapperspb.StringValue
	if req.Partner != "" {
		partner = wrapperspb.String(req.Partner)
	}
	if req.Note != "" {
		note = wrapperspb.String(req.Note)
	}

	_, err := h.userSrv.UserClient.Update(c.Request().Context(), &userpb.UpdateUserReq{
		ID:            req.ID,
		Name:          req.Name,
		Phone:         req.Phone,
		Partner:       partner,
		Note:          note,
		UpdatedUserID: handlers.GetUserInfo(c).ID,
	})
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.NotFound {
			return xerr.Get().NewErr(supports.GetErrMsg(supports.MissingPKError))
		}
		return err
	}
	return c.JSON(http.StatusOK, rsp)
}

// ChangePWDReq 用户修改密码入参
type ChangePWDReq struct {
	ID          uint32 `json:"id" validate:"required" example:"1"`
	OldPassword string `json:"old_password" validate:"required" example:"123456"`
	NewPassword string `json:"new_password" validate:"required" example:"1234567"`
}

func (p *ChangePWDReq) valid() error {
	if p.ID == 0 {
		return xerr.Get().NewErr(supports.GetErrMsg(supports.MissingPKError))
	}
	if p.OldPassword == p.NewPassword {
		return xerr.Get().NewErr(supports.GetErrMsg(supports.SamePasswordError))
	}
	return nil
}

// ChangePWD 修改密码
func (h *UserHandler) ChangePWD(c echo.Context) error {
	var (
		req ChangePWDReq
		rsp handlers.RspRow
	)
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := req.valid(); err != nil {
		return err
	}
	_, err := h.userSrv.UserClient.ChangePWD(c.Request().Context(), &userpb.ChangePWDReq{
		ID:            req.ID,
		OldPassword:   req.OldPassword,
		NewPassword:   req.NewPassword,
		UpdatedUserID: handlers.GetUserInfo(c).ID,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, rsp)
}

// Delete 删除用户
func (h *UserHandler) Delete(c echo.Context) error {
	var (
		req handlers.Uint32IDReq
		rsp handlers.RspRow
	)
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := req.Valid(); err != nil {
		return err
	}
	_, err := h.userSrv.UserClient.Delete(c.Request().Context(), &userpb.DeleteUserReq{
		ID:            req.ID,
		DeletedUserID: handlers.GetUserInfo(c).ID,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, rsp)
}
