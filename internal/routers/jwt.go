// Package routers HTTP 路由及中间件
package routers

import (
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/Jedeft/demo-micro-gw-admin/internal/handlers"
)

// jwtMiddleware JWT 认证中间件（替代已从 echo 核心中移除的 middleware.JWTWithConfig）
func jwtMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 跳过不需要认证的路径
			if handlers.AuthSkipper(c) {
				return next(c)
			}

			// 从 Header 中提取 token
			auth := c.Request().Header.Get("AuthToken")
			auth = strings.TrimPrefix(auth, "Bearer ")
			if auth == "" {
				return handlers.InvalidJWT(nil, c)
			}

			// 解析并校验 token
			_, err := handlers.ParseToken(auth, c)
			if err != nil {
				return handlers.InvalidJWT(err, c)
			}

			return next(c)
		}
	}
}
