# 单元测试生成设计

- 日期: 2026-07-03
- 范围: `demo-micro-gw-admin` 全部可测层
- 目标: 按 AGENTS.md 测试规范补齐缺失的单元测试

## 1. 背景与现状

项目当前无任何 `_test.go` 文件，go.mod 未引入测试依赖。AGENTS.md 要求：
- 文件命名 `*_test.go`，与被测文件同包
- 断言用 `testify/assert` 与 `testify/require`
- 外部依赖用 `gomock` 生成 Mock
- services 层行覆盖率不低于 90%
- 新增 API 端点必须同时提供单元测试与集成测试

## 2. 测试架构决策

### 2.1 测试工具链
- `github.com/stretchr/testify`：assert/require 断言
- `go.uber.org/mock`：gomock 运行时
- `github.com/alicebob/miniredis/v2`：redis 假服务器，覆盖 auth/jwt 的 redis 读写路径且不修改 handlers

### 2.2 Mock 生成（反射模式）
- 指令：`mockgen github.com/Jedeft/demo-micro-base-user/api/protobuf UserClient`
- `internal/mocks/gen.go` 放置 `//go:generate` 指令，执行 `go generate ./...` 重生成
- 反射模式直接反射外部模块的真实 `UserClient` 接口，外部加方法时自动同步

### 2.3 测试基础设施布局
```
internal/
  mocks/
    gen.go                  # //go:generate 指令
    mock_UserClient.go      # gomock 生成
  testutil/
    redis.go                # miniredis 启动 + cache.Init 注册
    echo.go                 # echo.Context / httptest 请求构造
    config.go               # configs.Config 测试值注入
```

### 2.4 redis 测试策略（不改 handlers）
`cache.Redis(alias)` 读取 xuanwu cache 包内未导出的 `sync.Map`，由 `cache.Init` 写入。
测试 TestMain 中调用 `cache.Init(cache.Config{Alias: supports.CacheJWT, Addr: miniredis.Addr()})`
注册一个指向 miniredis 的客户端。生产代码零改动，不触碰 handlers 红线。

## 3. 生产代码接缝（唯一非测试改动，不在 handlers 红线内）

`internal/services/user_service.go` 的 `InitUser` 当前直接调用 `grpc.GetConn("demo-base-user")`
（依赖 DNS 解析，无法单测）。引入 `var` 接缝使其可测：

```go
var newUserClient = func() (userpb.UserClient, error) {
    conn, err := grpc.GetConn("demo-base-user")
    if err != nil {
        return nil, err
    }
    return userpb.NewUserClient(conn), nil
}

func InitUser() {
    client, err := newUserClient()
    if err != nil {
        log.Get().Bg().Fatal("init user service: get grpc conn failed", zap.Error(err))
    }
    UserService = UserSrv{UserClient: client}
}
```

测试替换 `newUserClient` 返回 mock，覆盖 `InitUser` 成功路径；Fatal 路径必要时用子进程测试兜底。

## 4. 逐文件测试计划

| 文件 | 测试要点 | 依赖 |
|------|---------|------|
| `supports/utils/remove_repeat.go` | 表驱动：空切片/全重复/部分重复，uint32 与 string 两函数 | 无 |
| `supports/error.go` | `GetErrMsg` 已知码返回对应串、未知码返回空 | 无 |
| `handlers/common.go` | `Uint32IDReq.Valid`(0/非0)、`StringIDReq.Valid`(空/非空) | xerr(零值可用) |
| `handlers/users/user.go` | 三个 `valid()`；Add/Get/List/Search/Update/ChangePWD/Delete 各 happy + gRPC 错误分支（NotFound→MissingPKError）；List 的 limit clamp(0/>100) 与 offset 负值 | gomock UserClient |
| `handlers/auth/authrization.go` | `LoginReq.valid`；Login happy + NotFound/PermissionDenied→UsernamePwdError + 其他系统错误；Logout happy + redis Del 失败 | gomock + miniredis |
| `handlers/jwt.go` | CreateToken；ParseToken 成功/签名方法错/redis 不存在→401；AuthSkipper 各路径；GetUserInfo；InvalidJWT | miniredis + configs |
| `routers/jwt.go` | 中间件：跳过路径/无 token→401/有效放行/无效→401 | miniredis + configs |
| `handlers/error.go` | ErrorHandler：bizErr/echo.HTTPError/默认(debug 开关)/已 Committed/HEAD | echo |
| `configs/config.go` | Init 加载临时 toml 成功+失败；RandomPort 置 Port="0" | 临时配置文件 |
| `cmd/cmd.go` | Parse 缺参→panic；正常解析赋值 Arg | os.Args 操控 |
| `routers/router.go` | New() 路由注册齐全 | configs 注入 |

## 5. 集成测试（build tag `integration`）

- `internal/grpc/conn_test.go`(`//go:build integration`)：本地 gRPC server 绑 `127.0.0.1:8080`
  （端口占用则 `t.Skip`），验证 `GetConn` 缓存同 name 返回同 conn、`CloseAll` 清理。
- `internal/application_test.go`(`//go:build integration`)：临时 toml + miniredis + 本地 gRPC server
  跑 Init→Start→Stop。

契合 AGENTS.md：`go test ./...`（单测）与 `go test -tags=integration ./...`（集成）分离。

## 6. handler 测试注入方式

`UserHandler.userSrv` 取自 `NewUserHandler` 中复制的 `services.UserService` 全局变量。
测试中先 `services.UserService.UserClient = mockClient`，再 `NewUserHandler()`，handler 即获得带 mock 的副本。
同理 `AuthrizationHandler`。零生产改动。

## 7. 验证与记录

- 实现/验证后运行：`go fmt ./... && go vet ./... && golangci-lint run && go test ./... -v -cover`，
  集成：`go test -tags=integration ./...`
- 按 AGENTS.md 写 `docs/process/2026-07-03-unit-tests-generation.md` 记录命令、覆盖率、改动文件。
- 目录结构新增（mocks/testutil）→ 同步更新 `README.md`。

## 8. 验收标准

- 全部 `go test ./...` 通过，services 包覆盖率 ≥ 90%
- `golangci-lint run` 无新增告警
- 无任何 handlers 包生产代码改动
- 仅在 `services/user_service.go` 引入一处 `var` 接缝
- README 与 docs/process 同步更新
