# admin 网关微服务

## 目录结构

```
project
    ├── cmd
    │   ├── server
    │   │   ├── server.toml
    │   │   ├── xuanwu.toml
    │   │   └── main.go（程序入口）
    │   └── cmd.go（程序入参定义）
    ├── deployments（部署依赖目录）
    │   └── Dockerfile
    ├── docs
    ├── internal（本工程内目录，此目录下.go文件外部无法引用）
    │   ├── configs
    │   │   └── config.go
    │   ├── grpc
    │   │   ├── breaker.go（熔断、降级操作）
    │   │   └── service.go（grpc client初始化）
    │   ├── handlers
    │   │   ├── auth（认证模块）
    │   │   ├── users（用户模块）
    │   │   ├── common.go（公共响应体）
    │   │   ├── error.go（错误处理）
    │   │   └── jwt.go（JWT 处理）
    │   ├── routers
    │   │   └── router.go（echo 资源初始化）
    │   ├── services
    │   │   ├── service.go（服务层初始化）
    │   │   └── user_service.go（用户服务层）
    │   ├── supports
    │   │   ├── utils（工程内工具包）
    │   │   ├── cache.go（缓存别名）
    │   │   └── error.go（服务内错误定义）
    │   └── application.go（应用实例）
    ├── .gitignore
    ├── .gitlab-ci.yml
    ├── .golangci.yml
    ├── go.mod
    ├── Makefile
    └── README.md
```

## 架构说明

本项目遵循整洁架构（Clean Architecture）原则，依赖方向为：

```
handlers → services → 下游业务服务 rpc 调用
```

- **handlers**: HTTP 接口实现处，主要负责参数校验
- **services**: 核心业务逻辑处理层，负责调用下游业务微服务
- **grpc**: gRPC 客户端初始化与服务发现

## 本地编译

```shell
make build
make run
```

## 代码检测

```shell
make golangci
```

## 镜像构建

```shell
make docker-build DOCKER_TAG=latest
```
