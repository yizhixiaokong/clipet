.PHONY: build dev run run-dev status feed play init clean fmt lint test validate help

# ── 构建 ──────────────────────────────────────────────
build:                ## 构建 clipet 和 clipet-dev
	go build -o clipet     ./cmd/clipet/
	go build -o clipet-dev ./cmd/clipet-dev/

clipet:               ## 仅构建 clipet
	go build -o clipet ./cmd/clipet/

dev:                  ## 仅构建 clipet-dev
	go build -o clipet-dev ./cmd/clipet-dev/

# ── 运行 ──────────────────────────────────────────────
run: clipet           ## 启动 TUI
	./clipet

init: clipet          ## 创建新宠物
	./clipet init

status: clipet        ## 查看宠物状态
	./clipet status

feed: clipet          ## 喂食
	./clipet feed

play: clipet          ## 玩耍
	./clipet play

# ── 开发工具 ──────────────────────────────────────────
validate: dev         ## 校验内置猫物种包
	./clipet-dev validate internal/assets/builtins/cat-pack

# ── 代码质量 ──────────────────────────────────────────
fmt:                  ## 格式化代码
	gofmt -w .

lint: fmt             ## 静态检查（需安装 staticcheck）
	@command -v staticcheck >/dev/null 2>&1 && staticcheck ./... || echo "staticcheck not installed, skipping"

test:                 ## 运行测试
	go test ./...

# ── 清理 ──────────────────────────────────────────────
clean:                ## 删除构建产物
	rm -f clipet clipet-dev

# ── 帮助 ──────────────────────────────────────────────
help:                 ## 显示帮助
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | awk -F ':.*## ' '{printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
