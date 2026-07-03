// Package testutil 提供跨包共享的测试辅助。
package testutil

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"

	"github.com/Jedeft/xuanwu/cache"

	"github.com/Jedeft/demo-micro-gw-admin/internal/supports"
)

// StartRedis 启动一个 miniredis 实例并注册到 xuanwu cache（alias=supports.CacheJWT）。
// 重复调用安全（覆盖旧 alias）。miniredis 在 t.Cleanup 时自动关闭。
func StartRedis(t *testing.T) *miniredis.Miniredis {
	t.Helper()
	mr := miniredis.RunT(t)
	require.NoError(t, cache.Init(cache.Config{
		Alias:        supports.CacheJWT,
		Addr:         mr.Addr(),
		ReadTimeout:  1,
		WriteTimeout: 1,
	}))
	return mr
}
