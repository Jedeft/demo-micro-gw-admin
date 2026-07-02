package internal

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/Jedeft/xuanwu"
	"github.com/labstack/echo/v4"

	"github.com/Jedeft/demo-micro-gw-admin/cmd"
	"github.com/Jedeft/demo-micro-gw-admin/internal/configs"
	"github.com/Jedeft/demo-micro-gw-admin/internal/grpc"
	"github.com/Jedeft/demo-micro-gw-admin/internal/routers"
	"github.com/Jedeft/demo-micro-gw-admin/internal/services"
)

// Application application
type Application struct {
	e *echo.Echo
}

// Start start
func (app *Application) Start() error {
	// Start server，注意此处不能阻塞，go协程出去处理，注册recover中间件后echo会处理好panic
	go func() {
		err := app.e.Start(":" + configs.Config.Server.Port)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
	return nil
}

// Reload reaload
func (app *Application) Reload() error { return nil }

// Stop stop
func (app *Application) Stop() error {
	// 此处和服务器超时时间保持一致，所有的连接都请求结束后再关闭相关资源
	ctx, cancel := context.WithTimeout(context.Background(), configs.Config.Server.HTTPTimeout*time.Second)
	defer cancel()
	if err := app.e.Shutdown(ctx); err != nil {
		grpc.CloseAll()
		_ = xuanwu.Destroy()
		return err
	}
	grpc.CloseAll()
	return xuanwu.Destroy()
}

// Init init
func (app *Application) Init() error {
	// config 加载
	if err := configs.Init(); err != nil {
		return err
	}
	// xuanwu库初始化
	if err := xuanwu.Init(); err != nil {
		return err
	}
	// services 层初始化-依赖 gRPC 连接
	services.Init()

	// routers 创建使用了xuanwu插件，注意初始化顺序
	app.e = routers.New()
	return nil
}

// Command 执行命令行参数
func (app *Application) Command() {
	cmd.Parse()
}
