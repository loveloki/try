#!/bin/sh
# try 安装脚本：安装 TUI（try）与 GUI（try-gui）
# 用法: curl -fsSL https://raw.githubusercontent.com/loveloki/try/main/install.sh | sh
#
# 环境变量：
#   TRY_INSTALL_DIR   安装目录（默认 ~/.local/bin）
#   TRY_INSTALL_GUI   设为 0 时跳过 try-gui（默认安装）
set -e

REPO="loveloki/try"
INSTALL_DIR="${TRY_INSTALL_DIR:-$HOME/.local/bin}"
INSTALL_GUI="${TRY_INSTALL_GUI:-1}"

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

download() {
    URL="$1"
    DEST="$2"
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$URL" -o "$DEST"
    else
        wget -q "$URL" -O "$DEST"
    fi
}

install_binary() {
    SRC="$1"
    NAME="$2"
    if [ ! -f "$SRC" ]; then
        return 1
    fi
    mv "$SRC" "${INSTALL_DIR}/${NAME}"
    chmod +x "${INSTALL_DIR}/${NAME}"
    return 0
}

install() {
    FILENAME="try_${OS}_${ARCH}.tar.gz"
    URL="https://github.com/${REPO}/releases/download/v${VERSION}/${FILENAME}"
    TMPDIR=$(mktemp -d)
    trap 'rm -rf "$TMPDIR"' EXIT

    echo "下载 try v${VERSION} (${OS}/${ARCH})..."
    download "$URL" "${TMPDIR}/${FILENAME}"
    tar -xzf "${TMPDIR}/${FILENAME}" -C "$TMPDIR"
    mkdir -p "$INSTALL_DIR"

    if ! install_binary "${TMPDIR}/try" "try"; then
        echo "归档中缺少 try 二进制" >&2
        exit 1
    fi
    echo ""
    echo "✓ try v${VERSION} 已安装到 ${INSTALL_DIR}/try"

    if [ "$INSTALL_GUI" != "0" ]; then
        if install_binary "${TMPDIR}/try-gui" "try-gui"; then
            echo "✓ try-gui v${VERSION} 已安装到 ${INSTALL_DIR}/try-gui"
        else
            echo ""
            echo "⚠  当前平台归档未包含 try-gui，已跳过 GUI 安装。"
            echo "   可从源码构建：CGO_ENABLED=1 go install github.com/loveloki/try/cmd/try-gui@latest"
        fi
    else
        echo ""
        echo "已跳过 try-gui（TRY_INSTALL_GUI=0）"
    fi

    case ":$PATH:" in
        *":${INSTALL_DIR}:"*) ;;
        *)
            echo ""
            echo "⚠  ${INSTALL_DIR} 不在 \$PATH 中，请添加："
            echo "   export PATH=\"${INSTALL_DIR}:\$PATH\""
            ;;
    esac

    echo ""
    "${INSTALL_DIR}/try" install

    echo ""
    echo "重启终端或 source 对应配置文件即可使用："
    echo "  try       # TUI 选择器"
    if [ -x "${INSTALL_DIR}/try-gui" ]; then
        echo "  try-gui   # 原生 GUI"
    fi
}

detect_platform
get_latest_version
install
