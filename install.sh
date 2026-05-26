#!/bin/sh
# try 安装脚本
# 用法: curl -fsSL https://raw.githubusercontent.com/xleine/try/main/install.sh | sh
set -e

REPO="xleine/try"
INSTALL_DIR="${TRY_INSTALL_DIR:-$HOME/.local/bin}"

detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        linux)  OS="linux" ;;
        darwin) OS="darwin" ;;
        *)      echo "不支持的操作系统: $OS" >&2; exit 1 ;;
    esac

    case "$ARCH" in
        x86_64|amd64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *)             echo "不支持的架构: $ARCH" >&2; exit 1 ;;
    esac
}

get_latest_version() {
    if command -v curl >/dev/null 2>&1; then
        VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
            | grep '"tag_name"' | sed -E 's/.*"v?([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        VERSION=$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" \
            | grep '"tag_name"' | sed -E 's/.*"v?([^"]+)".*/\1/')
    else
        echo "需要 curl 或 wget" >&2; exit 1
    fi

    if [ -z "$VERSION" ]; then
        echo "无法获取最新版本号" >&2; exit 1
    fi
}

install() {
    FILENAME="try_${OS}_${ARCH}.tar.gz"
    URL="https://github.com/${REPO}/releases/download/v${VERSION}/${FILENAME}"
    TMPDIR=$(mktemp -d)
    trap 'rm -rf "$TMPDIR"' EXIT

    echo "下载 try v${VERSION} (${OS}/${ARCH})..."

    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$URL" -o "${TMPDIR}/${FILENAME}"
    else
        wget -q "$URL" -O "${TMPDIR}/${FILENAME}"
    fi

    tar -xzf "${TMPDIR}/${FILENAME}" -C "$TMPDIR"

    mkdir -p "$INSTALL_DIR"
    mv "${TMPDIR}/try" "${INSTALL_DIR}/try"
    chmod +x "${INSTALL_DIR}/try"

    echo ""
    echo "✓ try v${VERSION} 已安装到 ${INSTALL_DIR}/try"

    # 检查 PATH 是否包含安装目录
    case ":$PATH:" in
        *":${INSTALL_DIR}:"*) ;;
        *)
            echo ""
            echo "⚠  ${INSTALL_DIR} 不在 \$PATH 中，请添加："
            echo "   export PATH=\"${INSTALL_DIR}:\$PATH\""
            ;;
    esac

    # 自动设置 Shell 集成
    echo ""
    "${INSTALL_DIR}/try" install

    echo ""
    echo "重启终端或 source 对应配置文件即可使用 try 命令。"
}

detect_platform
get_latest_version
install
