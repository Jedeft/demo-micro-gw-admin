package testutil

import (
	"time"

	"github.com/Jedeft/demo-micro-gw-admin/internal/configs"
)

// SetTestConfig 注入一组可用于测试的 configs.Config 值，覆盖被测字段。
// 多次调用幂等，确保各测试看到一致的配置。
func SetTestConfig() {
	configs.Config.Server.Port = "0"
	configs.Config.Server.JWTSecret = "test-secret"
	configs.Config.Server.JWTTimeout = 3600 * time.Second
	configs.Config.Server.HTTPTimeout = 5 * time.Second
	configs.Config.Service.DemoBaseUser = "demo-base-user"
	configs.Config.Debug.IsFilterCaptcha = true
	configs.Config.Debug.RandomPort = false
	configs.Config.Debug.IsFilterToken = false
	configs.Config.Debug.EchoDebug = true
	configs.Config.Debug.DefaultUser.ID = 1
	configs.Config.Debug.DefaultUser.Username = "tester"
	configs.Config.Debug.DefaultUser.Name = "Tester"
}
