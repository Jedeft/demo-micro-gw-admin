package auth

import (
	"net/http"
	"time"

	"github.com/Jedeft/xuanwu/cache"
	"github.com/Jedeft/xuanwu/log"
	"github.com/Jedeft/xuanwu/xerr"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Jedeft/demo-micro-gw-admin/internal/configs"
	"github.com/Jedeft/demo-micro-gw-admin/internal/handlers"
	"github.com/Jedeft/demo-micro-gw-admin/internal/services"
	"github.com/Jedeft/demo-micro-gw-admin/internal/supports"
)

// AuthrizationHandler 用户handler
type AuthrizationHandler struct {
	log     log.Factory
	userSrv services.UserSrv
}

// NewAuthrizationHandler new AuthrizationHandler
func NewAuthrizationHandler() *AuthrizationHandler {
	return &AuthrizationHandler{
		log:     log.Get(),
		userSrv: services.UserService,
	}
}

// LoginReq 登录参数
type LoginReq struct {
	Username string `json:"username" validate:"required" example:"admin"`  // 用户名
	Password string `json:"password" validate:"required" example:"123456"` // 密码
	Captcha  string `json:"captcha" validate:"required" example:"64dfr7"`  // 验证码
}

// valid 校验
func (p *LoginReq) valid() *xerr.Err {
	if len(p.Username) == 0 {
		return xerr.Get().NewErr(supports.GetErrMsg(supports.UsernamePwdError))
	}
	if len(p.Password) == 0 {
		return xerr.Get().NewErr(supports.GetErrMsg(supports.UsernamePwdError))
	}
	if !configs.Config.Debug.IsFilterCaptcha {
		// TODO: 缓存中查询验证码比对
		if len(p.Captcha) == 0 {
			return xerr.Get().NewErr(supports.GetErrMsg(supports.InvalidCaptchaError))
		}
	}

	return nil
}

// LoginRsp 登录返回参数
type LoginRsp struct {
	Token     string `json:"token" example:"eyJhbGciOiJ"`     // token信息
	ExpiresAt int64  `json:"expires_at" example:"1640237748"` // token到期时间
}

// Login 用户登陆
func (h *AuthrizationHandler) Login(c echo.Context) error {
	var (
		req LoginReq
		rsp handlers.RspRow
	)
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := req.valid(); err != nil {
		return err
	}

	out, err := h.userSrv.Login(c.Request().Context(), req.Username, req.Password, c.RealIP())
	if err != nil {
		// 将 gRPC 状态码转换为业务错误
		st := status.Convert(err)
		switch st.Code() {
		case codes.NotFound, codes.PermissionDenied:
			return xerr.Get().NewErr(supports.GetErrMsg(supports.UsernamePwdError))
		default:
			h.log.Bg().Error("user login system error", zap.Error(err))
			return err
		}
	}

	// 将用户信息存储入 redis
	rc := cache.Redis(supports.CacheJWT)
	cs := rc.Set(c.Request().Context(), req.Username, true, time.Second*configs.Config.Server.JWTTimeout)
	if cs.Err() != nil {
		return cs.Err()
	}

	payload := &handlers.AuthPlyLoad{
		User: &handlers.UserInfo{
			ID:       out.ID,
			Username: out.Username,
			Name:     out.Name,
		},
	}
	token, err := handlers.CreateToken(payload)
	if err != nil {
		return err
	}
	rsp.Row = &LoginRsp{
		Token:     token,
		ExpiresAt: payload.ExpiresAt,
	}
	return c.JSON(http.StatusOK, rsp)
}

// Logout 用户注销
func (h *AuthrizationHandler) Logout(c echo.Context) error {
	var (
		rsp handlers.RspRow
	)

	// 从redis中去除jwt信息
	rc := cache.Redis(supports.CacheJWT)
	if err := rc.Del(c.Request().Context(), handlers.GetUserInfo(c).Username).Err(); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, rsp)
}
