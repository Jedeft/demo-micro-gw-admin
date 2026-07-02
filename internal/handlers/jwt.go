package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"

	"github.com/Jedeft/xuanwu/cache"

	"github.com/Jedeft/demo-micro-gw-admin/internal/configs"
	"github.com/Jedeft/demo-micro-gw-admin/internal/supports"
)

const (
	// ParsedUserKey 解析后的用户Key
	ParsedUserKey = "parsed_user"
)

// UserInfo 用户信息
type UserInfo struct {
	ID       uint32 `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

// AuthPlyLoad 用户信息添加到jwt中
type AuthPlyLoad struct {
	User *UserInfo
	jwt.StandardClaims
}

// InvalidJWT 不合法的JWT
func InvalidJWT(_ error, _ echo.Context) error {
	// err 与 c 均不使用：一律返回 401，错误细节交由 ErrorHandler 统一处理
	return echo.NewHTTPError(http.StatusUnauthorized)
}

// CreateToken create token
func CreateToken(claims *AuthPlyLoad) (signedToken string, err error) {
	claims.ExpiresAt = time.Now().Add(time.Second * configs.Config.Server.JWTTimeout).Unix()
	// 获取token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(configs.Config.Server.JWTSecret))
}

// ParseToken Token解析，解析后里面包含token的redis有状态校验以及存储
func ParseToken(auth string, c echo.Context) (interface{}, error) {
	// 解析 token
	token, err := jwt.ParseWithClaims(auth, &AuthPlyLoad{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected login method %v", token.Header["alg"])
		}
		return []byte(configs.Config.Server.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if token == nil {
		return nil, fmt.Errorf("token is error %v", token)
	}
	// 验证 token
	claims, ok := token.Claims.(*AuthPlyLoad)
	if !ok {
		return nil, fmt.Errorf("token type error %v", token)
	}

	// TODO: 从Redis中校验权限是否合法，也可以在这里限制用户单点登录
	rc := cache.Redis(supports.CacheJWT)
	result, err := rc.Exists(c.Request().Context(), claims.User.Username).Result()
	if err != nil {
		return nil, err
	}
	if result == 0 {
		return nil, echo.NewHTTPError(http.StatusUnauthorized)
	}

	// 解析后的token写入Context中
	c.Set(ParsedUserKey, claims.User)
	return token, nil
}

// AuthSkipper jwt忽略函数
func AuthSkipper(c echo.Context) bool {
	if configs.Config.Debug.IsFilterToken {
		// 若过滤Token 在这里给定默认账户信息
		c.Set(ParsedUserKey, &UserInfo{
			ID:       configs.Config.Debug.DefaultUser.ID,
			Name:     configs.Config.Debug.DefaultUser.Name,
			Username: configs.Config.Debug.DefaultUser.Username,
		})
		return true
	}
	path := c.Path()
	switch path {
	case "/health/check":
		return true
	case "/v1/login":
		// 忽略登录接口
		return true
	}
	return false
}

// GetUserInfo 获取用户信息
func GetUserInfo(c echo.Context) *UserInfo {
	return c.Get(ParsedUserKey).(*UserInfo)
}
