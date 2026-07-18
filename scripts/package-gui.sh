#!/usr/bin/env bash
# 使用 fyne package 将已构建的 try-gui 打成平台官方包，写入 dist/try-gui_<os>_<arch>.*
# 用法: scripts/package-gui.sh <goos> <goarch> <version>
# 前置: dist/try-gui 或 dist/try-gui.exe 已由 go build 产出
set -euo pipefail

GOOS="${1:?goos required}"
GOARCH="${2:?goarch required}"
VERSION="${3:?version required}"

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SRC="${ROOT}/cmd/try-gui"
DIST="${ROOT}/dist"
ICON="${SRC}/Icon.png"

EXT=""
if [ "$GOOS" = "windows" ]; then
  EXT=".exe"
fi
BIN="${DIST}/try-gui${EXT}"

if [ ! -f "$ICON" ]; then
  echo "缺少 ${ICON}" >&2
  exit 1
fi
if [ ! -f "$BIN" ]; then
  echo "缺少预构建二进制 ${BIN}，请先 go build ./cmd/try-gui" >&2
  exit 1
fi

command -v fyne >/dev/null 2>&1 || {
  echo "未找到 fyne CLI，请先: go install fyne.io/tools/cmd/fyne@latest" >&2
  exit 1
}

WORKDIR="$(mktemp -d)"
trap 'rm -rf "$WORKDIR"' EXIT

cp -R "$SRC/." "$WORKDIR/"
cp "$BIN" "${WORKDIR}/try-gui${EXT}"
cd "$WORKDIR"

if command -v sed >/dev/null 2>&1; then
  sed -i.bak "s/^Version = .*/Version = \"${VERSION}\"/" FyneApp.toml
  rm -f FyneApp.toml.bak
fi

# 使用已构建二进制，避免 fyne 二次编译破坏 GOARCH 交叉构建结果
fyne package \
  --os "$GOOS" \
  --icon Icon.png \
  --app-id com.loveloki.try.gui \
  --app-version "$VERSION" \
  --name Try \
  --release \
  --executable "./try-gui${EXT}"

OUT_BASE="try-gui_${GOOS}_${GOARCH}"

case "$GOOS" in
  darwin)
    APP=""
    for cand in Try.app try-gui.app; do
      if [ -d "$cand" ]; then
        APP="$cand"
        break
      fi
    done
    if [ -z "$APP" ]; then
      echo "fyne package 未产出 .app" >&2
      ls -la >&2
      exit 1
    fi
    if [ "$APP" != "Try.app" ]; then
      mv "$APP" Try.app
    fi
    if command -v zip >/dev/null 2>&1; then
      (cd "$WORKDIR" && zip -qry "${DIST}/${OUT_BASE}.app.zip" Try.app)
    else
      (cd "$WORKDIR" && tar -a -cf "${DIST}/${OUT_BASE}.app.zip" Try.app)
    fi
    ;;
  windows)
    EXE=""
    for cand in Try.exe try-gui.exe; do
      if [ -f "$cand" ]; then
        EXE="$cand"
        break
      fi
    done
    if [ -z "$EXE" ]; then
      echo "fyne package 未产出 Windows exe" >&2
      ls -la >&2
      exit 1
    fi
    if [ "$EXE" != "try-gui.exe" ]; then
      mv "$EXE" try-gui.exe
    fi
    # Windows runner 通常无 zip，用 tar -a 生成 zip
    if command -v zip >/dev/null 2>&1; then
      (cd "$WORKDIR" && zip -qry "${DIST}/${OUT_BASE}.zip" try-gui.exe)
    else
      (cd "$WORKDIR" && tar -a -cf "${DIST}/${OUT_BASE}.zip" try-gui.exe)
    fi
    ;;
  linux)
    if [ -d usr ]; then
      tar -czf "${DIST}/${OUT_BASE}.tar.gz" usr
    else
      TAR=""
      for cand in Try.tar.gz Try.tar.xz try.tar.gz; do
        if [ -f "$cand" ]; then
          TAR="$cand"
          break
        fi
      done
      if [ -z "$TAR" ]; then
        echo "fyne package 未产出 Linux 安装树" >&2
        ls -la >&2
        exit 1
      fi
      case "$TAR" in
        *.tar.xz) xz -dc "$TAR" | gzip -c > "${DIST}/${OUT_BASE}.tar.gz" ;;
        *) cp "$TAR" "${DIST}/${OUT_BASE}.tar.gz" ;;
      esac
    fi
    ;;
  *)
    echo "不支持的 GOOS: $GOOS" >&2
    exit 1
    ;;
esac

ls -la "${DIST}/${OUT_BASE}".*
echo "✓ GUI 官方包: ${DIST}/${OUT_BASE}.*"
