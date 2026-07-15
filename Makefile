# 取消 GOROOT 环境变量导出，避免系统 /usr/local/go(1.25.4) 与 GOTOOLCHAIN 自动
# 下载的 toolchain(1.25.8) 冲突导致 compile 版本不匹配。现代 Go 会自动推断 GOROOT。
unexport GOROOT

# 镜像相关变量（可通过环境变量覆盖）
DOCKER_TARGET ?= demo-micro-gw-admin
DOCKER_TAG    ?= latest



.PHONY: build build-linux run docker-build docker-release golangci

# 本地构建
build:
	CGO_ENABLED=0 go build -ldflags=-checklinkname=0 -trimpath -o app ./cmd/server

# 交叉编译 linux/amd64 二进制（供 Dockerfile COPY，避免镜像内拉取私有仓库）
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags=-checklinkname=0 -trimpath -o app ./cmd/server

run:
	go run -ldflags=-checklinkname=0 ./cmd/server -sc ./cmd/server/server.toml -hc ./cmd/server/xuanwu.toml

# 镜像构建（公有镜像，无需 GitLab 凭证）
docker-build: build-linux
	docker build -f deployments/Dockerfile -t $(DOCKER_TARGET):$(DOCKER_TAG) .

# 镜像推送
docker-release:
	# latest 指向当前tag,全部推送
	docker tag 	$(DOCKER_TARGET):$(DOCKER_TAG) $(DOCKER_TARGET)
	docker push 	$(DOCKER_TARGET):$(DOCKER_TAG)
	docker push 	$(DOCKER_TARGET)

# CI检测
golangci:
	golangci-lint run
