#!/usr/bin/env bash
# 使用 fyne package 将 try-gui 打成平台官方包，写入 dist/try-gui_<os>_<arch>.*
# 用法: scripts/package-gui.sh <goos> <goarch> <version>
# 前置: dist/try-gui[.exe] 已由 go build 产出（Windows 在 fyne 失败时作回退）
set -euo pipefail

GOOS="${1:?goos required}"
GOARCH="${2:?goarch required}"
VERSION="${3:?version required}"

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SRC="${ROOT}/cmd/try-gui"
DIST="${ROOT}/dist"
ICON="${SRC}/Icon.png"
OUT_BASE="try-gui_${GOOS}_${GOARCH}"

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

mkdir -p "$DIST"

make_zip() {
  local src="$1"
  local dest="$2"
  local parent
  parent="$(dirname "$src")"
  local base
  base="$(basename "$src")"
  if command -v zip >/dev/null 2>&1; then
    (cd "$parent" && zip -qry "$dest" "$base")
  else
    (cd "$parent" && tar -a -cf "$dest" "$base")
  fi
}

package_unix_like() {
  # darwin/linux：用已构建二进制 + 临时目录，避免二次交叉编译
  local workdir
  workdir="$(mktemp -d)"
  trap 'rm -rf "$workdir"' RETURN

  cp -R "$SRC/." "$workdir/"
  cp "$BIN" "${workdir}/try-gui${EXT}"
  (
    cd "$workdir"
    if command -v sed >/dev/null 2>&1; then
      sed -i.bak "s/^Version = .*/Version = \"${VERSION}\"/" FyneApp.toml
      rm -f FyneApp.toml.bak
    fi
    fyne package \
      --os "$GOOS" \
      --icon Icon.png \
      --app-id com.loveloki.try.gui \
      --app-version "$VERSION" \
      --name Try \
      --release \
      --executable "./try-gui${EXT}"
  )

  case "$GOOS" in
    darwin)
      local app=""
      for cand in "${workdir}/Try.app" "${workdir}/try-gui.app"; do
        if [ -d "$cand" ]; then
          app="$cand"
          break
        fi
      done
      if [ -z "$app" ]; then
        echo "fyne package 未产出 .app" >&2
        ls -la "$workdir" >&2
        exit 1
      fi
      if [ "$(basename "$app")" != "Try.app" ]; then
        mv "$app" "${workdir}/Try.app"
        app="${workdir}/Try.app"
      fi
      make_zip "$app" "${DIST}/${OUT_BASE}.app.zip"
      ;;
    linux)
      if [ -d "${workdir}/usr" ]; then
        tar -czf "${DIST}/${OUT_BASE}.tar.gz" -C "$workdir" usr
      else
        local tarfile=""
        for cand in Try.tar.gz Try.tar.xz try.tar.gz; do
          if [ -f "${workdir}/${cand}" ]; then
            tarfile="${workdir}/${cand}"
            break
          fi
        done
        if [ -z "$tarfile" ]; then
          echo "fyne package 未产出 Linux 安装树" >&2
          ls -la "$workdir" >&2
          exit 1
        fi
        case "$tarfile" in
          *.tar.xz) xz -dc "$tarfile" | gzip -c > "${DIST}/${OUT_BASE}.tar.gz" ;;
          *) cp "$tarfile" "${DIST}/${OUT_BASE}.tar.gz" ;;
        esac
      fi
      ;;
  esac
}

package_windows() {
  # Windows：在仓库根目录从源码打包（需 go.mod + MinGW），不把二进制拷进无模块的临时目录。
  # fyne --executable 在部分 runner 上嵌入图标会失败；失败则回退为已构建的 windowsgui exe。
  local out_exe="${ROOT}/Try.exe"
  rm -f "$out_exe" "${ROOT}/try-gui.exe"

  set +e
  (
    cd "$ROOT"
    export CGO_ENABLED=1
    fyne package \
      --os windows \
      --source-dir ./cmd/try-gui \
      --icon ./cmd/try-gui/Icon.png \
      --app-id com.loveloki.try.gui \
      --app-version "$VERSION" \
      --name Try \
      --release
  )
  local rc=$?
  set -e

  local packaged=""
  for cand in "${ROOT}/Try.exe" "${ROOT}/try-gui.exe" "${SRC}/Try.exe"; do
    if [ -f "$cand" ]; then
      packaged="$cand"
      break
    fi
  done

  if [ "$rc" -ne 0 ] || [ -z "$packaged" ]; then
    echo "⚠  fyne package 失败或未产出 exe (exit=${rc})，回退为预构建 dist/try-gui.exe" >&2
    ls -la "$ROOT"/*.exe 2>/dev/null || true
    packaged="$BIN"
  fi

  local stage="${DIST}/.gui-stage"
  rm -rf "$stage"
  mkdir -p "$stage"
  cp "$packaged" "${stage}/try-gui.exe"
  make_zip "${stage}/try-gui.exe" "${DIST}/${OUT_BASE}.zip"
  rm -rf "$stage"
  rm -f "$out_exe" "${ROOT}/try-gui.exe" "${SRC}/Try.exe"
}

case "$GOOS" in
  darwin|linux) package_unix_like ;;
  windows) package_windows ;;
  *)
    echo "不支持的 GOOS: $GOOS" >&2
    exit 1
    ;;
esac

ls -la "${DIST}/${OUT_BASE}".*
echo "✓ GUI 官方包: ${DIST}/${OUT_BASE}.*"
