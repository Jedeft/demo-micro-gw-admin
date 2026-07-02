# Dockerfile References: https://docs.docker.com/engine/reference/builder/
#
# HTTP 网关运行镜像。
# 二进制（app）均在宿主机预置，镜像内不执行任何网络下载，
# 规避容器内拉取 github.com / 私有仓库失败的问题。

FROM debian:bookworm-slim

WORKDIR /svr

# 本地预编译应用二进制（由 Makefile docker-build -> build-linux 生成）
COPY app .

# 与 cmd/server/server.toml 中 [server] port = "8080" 对齐
EXPOSE 8080

CMD ["./app", "application"]
