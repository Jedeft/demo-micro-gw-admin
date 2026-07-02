// Package grpc gRPC 连接管理
package grpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Jedeft/xuanwu/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

var (
	connMu  sync.Mutex
	connMap = make(map[string]*grpc.ClientConn)
)

const (
	defaultPort = "8080"
)

// GetConn 获取或创建到指定服务的 gRPC 连接
// name 为服务注册名（如 "demo-base-user"），在 K8s 环境中通过 DNS 解析到实际地址
//
// 线程安全：内部使用 sync.Mutex 保护并发创建。
func GetConn(name string) (*grpc.ClientConn, error) {
	connMu.Lock()
	defer connMu.Unlock()

	if conn, ok := connMap[name]; ok {
		return conn, nil
	}

	addr := fmt.Sprintf("%s:%s", name, defaultPort)
	logger := log.Get().Bg().With(zap.String("service", name), zap.String("addr", addr))

	// NewClient 是非阻塞的，仅在校验 DialOption 参数失败时返回错误。
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(1024*1024*10), // 10MB
		),
	)
	if err != nil {
		logger.Error("grpc new client error", zap.Error(err))
		return nil, fmt.Errorf("grpc new client %s: %w", addr, err)
	}

	// NewClient 不会阻塞等待连接建立，此处主动 Connect 并在超时内轮询连接状态，
	// 复刻原 grpc.DialContext + WithBlock 的"服务不可达即快速失败"行为。
	dialCtx, dialCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer dialCancel()
	conn.Connect()
	for {
		state := conn.GetState()
		if state == connectivity.Ready {
			break
		}
		if !conn.WaitForStateChange(dialCtx, state) {
			break // ctx 超时或取消
		}
	}
	if conn.GetState() != connectivity.Ready {
		_ = conn.Close()
		logger.Error("grpc connection timeout")
		return nil, fmt.Errorf("grpc dial %s: connection not ready within timeout", addr)
	}

	connMap[name] = conn
	logger.Info("grpc connection established")
	return conn, nil
}

// CloseAll 关闭所有 gRPC 连接（用于优雅关闭）
func CloseAll() {
	connMu.Lock()
	defer connMu.Unlock()

	for name, conn := range connMap {
		if err := conn.Close(); err != nil {
			log.Get().Bg().Error("grpc close error",
				zap.String("service", name),
				zap.Error(err),
			)
		}
		delete(connMap, name)
	}
}
