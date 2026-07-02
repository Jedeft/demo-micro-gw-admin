package cmd

import (
	"flag"
	"os"
)

const (
	// FlagsServerConfig 服务 config 参数
	FlagsServerConfig = iota
	// FlagsXuanwuConfig xuanwu config 参数
	FlagsXuanwuConfig
)

// Flag 命令参数配置
type Flag struct {
	Name  string
	Value string
	Usage string
	Arg   string
}

// DefaultFlags 默认命令参数集合
var DefaultFlags = map[int]*Flag{
	FlagsServerConfig: {
		Name:  "sc",
		Value: "./server.toml",
		Usage: "server config path",
	},
	FlagsXuanwuConfig: {
		Name:  "hc",
		Value: "./xuanwu.toml",
		Usage: "xuanwu config path",
	},
}

// Parse 命令解析
func Parse() {
	if len(os.Args) < 2 {
		panic("missing app flag arg")
	}
	// 防止flag全局包与micro发生冲突，此处由os.Args第一个参数作为本服务cmd入口标志
	set := flag.NewFlagSet(os.Args[1], flag.PanicOnError)
	for _, v := range DefaultFlags {
		set.StringVar(&v.Arg, v.Name, v.Value, v.Usage)
	}
	_ = set.Parse(os.Args[2:])
}
