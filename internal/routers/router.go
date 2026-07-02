package routers

import (
	"time"

	hh "github.com/Jedeft/xuanwu/pkg/handler/http"
	"github.com/Jedeft/xuanwu/tracer"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/Jedeft/demo-micro-gw-admin/internal/configs"
	"github.com/Jedeft/demo-micro-gw-admin/internal/handlers"
	"github.com/Jedeft/demo-micro-gw-admin/internal/handlers/auth"
	"github.com/Jedeft/demo-micro-gw-admin/internal/handlers/users"
)

// New http router register.
func New() *echo.Echo {
	e := echo.New()
	e.Server.ReadTimeout = configs.Config.Server.HTTPTimeout * time.Second
	e.Server.WriteTimeout = configs.Config.Server.HTTPTimeout * time.Second
	e.HideBanner = true
	e.Debug = configs.Config.Debug.EchoDebug

	// 自定义错误handler
	e.HTTPErrorHandler = handlers.ErrorHandler

	// 路由权限校验
	e.Use(jwtMiddleware())
	// 访问日志：echo v4.15 弃用 middleware.Logger，改用 RequestLogger + LogValuesFunc
	// 自定义输出，写入 echo 自带 logger（与原行为一致，结构化 JSON）。
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogMethod:   true,
		LogURI:      true,
		LogStatus:   true,
		LogLatency:  true,
		LogRemoteIP: true,
		LogError:    true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			fields := map[string]any{
				"remote_ip": v.RemoteIP,
				"method":    v.Method,
				"uri":       v.URI,
				"status":    v.Status,
				"latency":   v.Latency.String(),
			}
			if v.Error != nil {
				fields["error"] = v.Error.Error()
				c.Logger().Errorj(fields)
				return nil
			}
			c.Logger().Infoj(fields)
			return nil
		},
	}))
	e.Use(middleware.Recover())
	e.Use(tracer.EchoMiddleware())

	// 健康检查
	hHandler := hh.NewHealthHandler()
	e.GET("/health/check", hHandler.Check)

	v1Group := e.Group("v1")

	authHandler := auth.NewAuthrizationHandler()
	v1Group.POST("/login", authHandler.Login)
	v1Group.POST("/logout", authHandler.Logout)

	userHandler := users.NewUserHandler()
	v1Group.POST("/user/add", userHandler.Add)
	v1Group.POST("/user/update", userHandler.Update)
	v1Group.POST("/user/password/update", userHandler.ChangePWD)
	v1Group.POST("/user/delete", userHandler.Delete)
	v1Group.GET("/user", userHandler.Get)
	v1Group.GET("/user/list", userHandler.List)
	v1Group.GET("/user/search", userHandler.Search)
	return e
}
