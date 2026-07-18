#!/bin/sh
# try 安装脚本：安装 TUI（try）与可选 GUI（官方 fyne package 产物）
# 用法: curl -fsSL https://raw.githubusercontent.com/loveloki/try/main/install.sh | sh
#
# 环境变量：
#   TRY_INSTALL_DIR   CLI 安装目录（默认 ~/.local/bin）
#   TRY_INSTALL_GUI   设为 0 时跳过 try-gui（默认安装）
#   TRY_APPS_DIR      macOS .app 安装目录（默认 ~/Applications）
set -e

REPO="loveloki/try"
INSTALL_DIR="${TRY_INSTALL_DIR:-$HOME/.local/bin}"
INSTALL_GUI="${TRY_INSTALL_GUI:-1}"
APPS_DIR="${TRY_APPS_DIR:-$HOME/Applications}"

detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        linux)  OS="linux" ;;
        darwin) OS="darwin" ;;
        *)      echo "不支持的操作系统: $OS（Windows 请使用 install.ps1）" >&2; exit 1 ;;
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

install_cli() {
    FILENAME="try_${OS}_${ARCH}.tar.gz"
    URL="https://github.com/${REPO}/releases/download/v${VERSION}/${FILENAME}"
    echo "下载 try v${VERSION} (${OS}/${ARCH})..."
    download "$URL" "${TMPDIR}/${FILENAME}"
    tar -xzf "${TMPDIR}/${FILENAME}" -C "$TMPDIR"
    mkdir -p "$INSTALL_DIR"

    if [ ! -f "${TMPDIR}/try" ]; then
        echo "归档中缺少 try 二进制" >&2
        exit 1
    fi
    mv "${TMPDIR}/try" "${INSTALL_DIR}/try"
    chmod +x "${INSTALL_DIR}/try"
    echo "✓ try v${VERSION} 已安装到 ${INSTALL_DIR}/try"
}

# Linux：将 fyne usr/local 树适配到 ~/.local
install_gui_linux() {
    PKG="try-gui_${OS}_${ARCH}.tar.gz"
    URL="https://github.com/${REPO}/releases/download/v${VERSION}/${PKG}"
    echo "下载 GUI 官方包 ${PKG}..."
    if ! download "$URL" "${TMPDIR}/${PKG}"; then
        echo "⚠  无 GUI 官方包（例如 linux/arm64），已跳过。" >&2
        return 1
    fi
    mkdir -p "${TMPDIR}/gui"
    tar -xzf "${TMPDIR}/${PKG}" -C "${TMPDIR}/gui"

    BIN=""
    for cand in \
        "${TMPDIR}/gui/usr/local/bin/Try" \
        "${TMPDIR}/gui/usr/local/bin/try-gui" \
        "${TMPDIR}/gui/usr/local/bin/try"
    do
        if [ -f "$cand" ]; then
            BIN="$cand"
            break
        fi
    done
    if [ -z "$BIN" ]; then
        BIN=$(find "${TMPDIR}/gui" -type f -perm -111 \( -name 'Try' -o -name 'try-gui' \) 2>/dev/null | head -1)
    fi
    if [ -z "$BIN" ] || [ ! -f "$BIN" ]; then
        echo "GUI 包中未找到可执行文件" >&2
        return 1
    fi

    mkdir -p "$INSTALL_DIR"
    mv "$BIN" "${INSTALL_DIR}/try-gui"
    chmod +x "${INSTALL_DIR}/try-gui"

    DESKTOP_SRC=$(find "${TMPDIR}/gui" -name '*.desktop' 2>/dev/null | head -1)
    APP_DIR="${HOME}/.local/share/applications"
    ICON_DIR="${HOME}/.local/share/icons/hicolor/256x256/apps"
    mkdir -p "$APP_DIR" "$ICON_DIR"

    ICON_SRC=$(find "${TMPDIR}/gui" \( -name 'Try.png' -o -name 'try.png' -o -name 'try-gui.png' \) 2>/dev/null | head -1)
    if [ -n "$ICON_SRC" ] && [ -f "$ICON_SRC" ]; then
        cp "$ICON_SRC" "${ICON_DIR}/try-gui.png"
        ICON_KEY="try-gui"
    else
        ICON_KEY="utilities-terminal"
    fi

    DESKTOP_OUT="${APP_DIR}/try-gui.desktop"
    if [ -n "$DESKTOP_SRC" ] && [ -f "$DESKTOP_SRC" ]; then
        # 改写 Exec / Icon 指向用户级路径
        sed \
            -e "s|^Exec=.*|Exec=${INSTALL_DIR}/try-gui|" \
            -e "s|^Icon=.*|Icon=${ICON_KEY}|" \
            "$DESKTOP_SRC" > "$DESKTOP_OUT"
    else
        cat > "$DESKTOP_OUT" <<EOF
[Desktop Entry]
Type=Application
Name=Try
Comment=Manage temporary experiment directories
Exec=${INSTALL_DIR}/try-gui
Icon=${ICON_KEY}
Terminal=false
Categories=Utility;
EOF
    fi
    # 确保 Name 为 Try，便于应用菜单搜索
    if ! grep -q '^Name=Try' "$DESKTOP_OUT" 2>/dev/null; then
        sed -i.bak 's/^Name=.*/Name=Try/' "$DESKTOP_OUT" 2>/dev/null || true
        rm -f "${DESKTOP_OUT}.bak"
    fi

    if command -v update-desktop-database >/dev/null 2>&1; then
        update-desktop-database "$APP_DIR" 2>/dev/null || true
    fi

    echo "✓ try-gui 已安装到 ${INSTALL_DIR}/try-gui"
    echo "✓ 桌面入口: ${DESKTOP_OUT}"
    return 0
}

install_gui_darwin() {
    PKG="try-gui_${OS}_${ARCH}.app.zip"
    URL="https://github.com/${REPO}/releases/download/v${VERSION}/${PKG}"
    echo "下载 GUI 官方包 ${PKG}..."
    if ! download "$URL" "${TMPDIR}/${PKG}"; then
        echo "⚠  无 GUI 官方包，已跳过。" >&2
        return 1
    fi
    mkdir -p "${TMPDIR}/gui" "$APPS_DIR"
    unzip -qo "${TMPDIR}/${PKG}" -d "${TMPDIR}/gui"
    if [ ! -d "${TMPDIR}/gui/Try.app" ]; then
        echo "GUI 包中缺少 Try.app" >&2
        return 1
    fi
    rm -rf "${APPS_DIR}/Try.app"
    mv "${TMPDIR}/gui/Try.app" "${APPS_DIR}/Try.app"

    # 同步 PATH：链到 .app 内二进制（fyne 可能命名为 Try 或 try-gui）
    APP_BIN=""
    for cand in \
        "${APPS_DIR}/Try.app/Contents/MacOS/try-gui" \
        "${APPS_DIR}/Try.app/Contents/MacOS/Try"
    do
        if [ -f "$cand" ]; then
            APP_BIN="$cand"
            break
        fi
    done
    if [ -z "$APP_BIN" ]; then
        APP_BIN=$(find "${APPS_DIR}/Try.app/Contents/MacOS" -type f -perm -111 2>/dev/null | head -1)
    fi
    if [ -n "$APP_BIN" ] && [ -f "$APP_BIN" ]; then
        mkdir -p "$INSTALL_DIR"
        ln -sfn "$APP_BIN" "${INSTALL_DIR}/try-gui"
    fi

    echo "✓ Try.app 已安装到 ${APPS_DIR}/Try.app"
    echo "  （首次打开若被拦截：右键 → 打开）"
    return 0
}

install_gui() {
    if [ "$INSTALL_GUI" = "0" ]; then
        echo ""
        echo "已跳过 try-gui（TRY_INSTALL_GUI=0）"
        return 0
    fi
    echo ""
    case "$OS" in
        linux)
            if ! install_gui_linux; then
                echo "   可手动下载 try-gui_${OS}_${ARCH}.tar.gz，或源码构建："
                echo "   CGO_ENABLED=1 go install github.com/loveloki/try/cmd/try-gui@latest"
            fi
            ;;
        darwin)
            if ! install_gui_darwin; then
                echo "   可手动下载 try-gui_${OS}_${ARCH}.app.zip 并拖入应用程序"
            fi
            ;;
    esac
}

install() {
    TMPDIR=$(mktemp -d)
    trap 'rm -rf "$TMPDIR"' EXIT

    install_cli
    install_gui

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
    if [ -x "${INSTALL_DIR}/try-gui" ] || [ -d "${APPS_DIR}/Try.app" ]; then
        echo "  try-gui   # 或从应用列表启动 Try"
    fi
}

detect_platform
get_latest_version
install
