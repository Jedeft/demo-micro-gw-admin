package configs

import (
	"fmt"
	"time"

	"github.com/Jedeft/xuanwu"
	"github.com/Jedeft/xuanwu/config"

	"github.com/Jedeft/demo-micro-gw-admin/cmd"
)

// Config 配置
var Config = struct {
	Server struct {
		Port        string        `toml:"port"`
		JWTSecret   string        `toml:"jwt_secret"`
		JWTTimeout  time.Duration `toml:"jwt_timeout"`
		HTTPTimeout time.Duration `toml:"http_timeout"`
	} `toml:"server"`

	Service struct {
		DemoBaseUser string `toml:"demo_base_user"`
	} `toml:"service"`

	// 调试参数
	Debug struct {
		IsFilterCaptcha bool `toml:"is_filter_captcha"` // 是否过滤验证码
		RandomPort      bool `toml:"random_port"`       // 是否随机端口
		IsFilterToken   bool `toml:"is_filter_token"`   // 是否过滤token
		EchoDebug       bool `toml:"echo_debug"`        // 是否开启Debug模式，Debug模式下http response会返回详细后台报错，仅用于本地调试使用
		DefaultUser     struct {
			ID       uint32 `toml:"id"`
			Username string `toml:"username"`
			Name     string `toml:"name"`
		} `toml:"default_user"`
	} `toml:"debug"`
}{}

// Init 配置初始化
func Init() error {
	// 初始化xuanwu
	if err := xuanwu.SetConfig(cmd.DefaultFlags[cmd.FlagsXuanwuConfig].Arg); err != nil {
		return fmt.Errorf("xuanwu config set err: %w", err)
	}

	if err := config.Load(cmd.DefaultFlags[cmd.FlagsServerConfig].Arg, &Config); err != nil {
		return fmt.Errorf("server config load err: %w", err)
	}

	// 开启随机端口，避免本地调试出现启动端口冲突
	if Config.Debug.RandomPort {
		Config.Server.Port = "0"
	}
	return nil
}
