#!/usr/bin/env bash
# 一键发布新版本：自动推断版本号、创建 tag 并推送触发 CI 构建
set -euo pipefail

# 将 GOPATH/bin 加入 PATH，确保 go install 安装的工具可被找到
export PATH="$(go env GOPATH)/bin:$PATH"

check_tool() {
    if ! command -v "$1" &>/dev/null; then
        echo "错误：未找到 $1，请先安装：$2"
        exit 1
    fi
}

check_tool svu "go install github.com/caarlos0/svu@latest"
check_tool git "https://git-scm.com"

# 检查工作区干净
if [ -n "$(git status --porcelain)" ]; then
    echo "错误：工作区有未提交的变更，请先提交或暂存"
    exit 1
fi

# 确保在 main 分支
branch=$(git branch --show-current)
if [ "$branch" != "main" ]; then
    echo "错误：当前在 $branch 分支，请切换到 main 分支"
    exit 1
fi

current=$(svu current)
next=$(svu next)
explicit=false

# 支持手动指定版本类型，或显式 tag（如 v0.4.0）
case "${1:-}" in
    patch) next=$(svu patch) ;;
    minor) next=$(svu minor) ;;
    major) next=$(svu major) ;;
    "")    ;;  # 默认自动推断
    v[0-9]*.[0-9]*.[0-9]*)
        next="$1"
        explicit=true
        ;;
    *)
        echo "用法: $0 [patch|minor|major|vX.Y.Z]"
        echo "  不带参数：根据 commit 历史自动推断版本"
        echo "  patch：补丁版本升级 (x.y.Z)"
        echo "  minor：次版本升级 (x.Y.0)"
        echo "  major：主版本升级 (X.0.0)"
        echo "  vX.Y.Z：显式指定版本号（跳过 svu）"
        exit 1
        ;;
esac

if [ "$explicit" = false ] && [ "$current" = "$next" ]; then
    echo "当前版本 $current 已是最新，没有需要发布的变更"
    exit 0
fi

echo "当前版本: $current"
echo "新版本:   $next"
echo ""

# 确认发布
read -rp "确认发布 $next？[y/N] " confirm
if [[ ! "$confirm" =~ ^[yY]$ ]]; then
    echo "已取消"
    exit 0
fi

# 构建并测试
echo ""
echo "构建..."
go build ./...
echo "构建通过 ✓"

echo "运行测试..."
go test ./... -timeout 60s
echo "测试通过 ✓"

# 创建 tag 并推送
echo ""
echo "创建 tag $next..."
git tag "$next"

if ! git push origin "$next"; then
    echo ""
    echo "错误：推送失败，回滚本地 tag"
    git tag -d "$next"
    exit 1
fi

echo ""
echo "✓ 已推送 $next，GitHub Actions 将自动构建并创建 Release"
