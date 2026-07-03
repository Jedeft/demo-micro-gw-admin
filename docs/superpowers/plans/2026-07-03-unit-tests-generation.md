# 单元测试生成 实现计划

> REQUIRED SUB-SKILL: superpowers:executing-plans（本会话内联执行）

Goal: 为 demo-micro-gw-admin 全部可测层补齐单元测试，services 覆盖率 ≥90%。
Architecture: internal/mocks + internal/testutil + gomock 反射 MockUserClient + miniredis；services 引入 var 接缝；grpc/application 走 //go:build integration。
Tech: Go 1.25.8 / testify / github.com/golang/mock v1.6.0 / alicebob/miniredis/v2 / echo httptest。

## Global Constraints
- 测试文件与被测文件同包；断言 testify assert+require；mock 用 golang/mock（与已装 mockgen 一致）。
- 红线：禁改 internal/handlers 生产代码；唯一生产改动在 internal/services/user_service.go。
- 提交须用户明确要求；conventional commit 无 AI 签名。每步后跑 fmt/vet/lint；目录变更同步 README；结果写 docs/process。

## Tasks
1. 测试基础设施：go.mod +testify +golang/mock +miniredis；internal/mocks(gen.go+生成 mock_UserClient.go)；internal/testutil(redis.go/echo.go/config.go)。
2. services：user_service.go 引入 var newUserClient 接缝；user_service_test.go 覆盖 Login 三路径 + InitUser 成功/Fatal。
3. supports：utils/remove_repeat_test.go 表驱动；supports/error_test.go GetErrMsg。
4. handlers：common_test.go(Valid) + jwt_test.go(CreateToken/ParseToken/AuthSkipper/GetUserInfo/InvalidJWT) + error_test.go(ErrorHandler)。
5. handlers/users：user_test.go(valid() + Add/Get/List/Search/Update/ChangePWD/Delete，mock 注入 via services.UserService.UserClient)。
6. handlers/auth：authrization_test.go(LoginReq.valid + Login/Logout，mock + miniredis)。
7. routers：jwt_test.go(jwtMiddleware) + router_test.go(New 路由)。
8. configs + cmd：config_test.go(Init+RandomPort 临时toml) + cmd_test.go(Parse panic/正常)。
9. 集成(build tag integration)：grpc/conn_integration_test.go + application_integration_test.go。
10. 验证：fmt/vet/lint/test -cover；更新 README；写 docs/process/2026-07-03-unit-tests-generation.md。

详见 spec: docs/superpowers/specs/2026-07-03-unit-tests-generation-design.md
