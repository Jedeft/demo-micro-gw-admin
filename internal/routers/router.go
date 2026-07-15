package routers

import (
	"time"

	"github.com/Jedeft/xuanwu/log"
	hh "github.com/Jedeft/xuanwu/pkg/handler/http"
	"github.com/Jedeft/xuanwu/tracer"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

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

	// tracer 最外层：先注入 span，后续中间件（含 access 日志）的 ctx 均可关联 trace_id。
	// otelecho 在 defer 中会恢复原始 ctx，故 tracer 必须在 RequestLogger 之外层，
	// 否则 LogValuesFunc 调用时 ctx 已被恢复、丢失 span。
	e.Use(tracer.EchoMiddleware())
	// 路由权限校验
	e.Use(jwtMiddleware())
	// 访问日志：echo v4.15 弃用 middleware.Logger，改用 RequestLogger + LogValuesFunc
	// 经 xuanwu access logger 写入 access.log，For(ctx) 关联 tracer 注入的 trace_id。
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogMethod:   true,
		LogURI:      true,
		LogStatus:   true,
		LogLatency:  true,
		LogRemoteIP: true,
		LogError:    true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			fields := []zap.Field{
				zap.String("remote_ip", v.RemoteIP),
				zap.String("method", v.Method),
				zap.String("uri", v.URI),
				zap.Int("status", v.Status),
				zap.String("latency", v.Latency.String()),
			}
			if v.Error != nil {
				log.Get("access").For(c.Request().Context()).Error("access", append(fields, zap.Error(v.Error))...)
				return nil
			}
			log.Get("access").For(c.Request().Context()).Info("access", fields...)
			return nil
		},
	}))
	e.Use(middleware.Recover())

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
