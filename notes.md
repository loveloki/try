---

## [2026-07-17] try-gui 对齐 TUI 首轮改造

- **本会话**：
  - 使用多 agent 方式分析 `try-gui/` 与现有 Bubbletea TUI 的差异，确认 `try-gui` 当前主要提供状态模型、导航 hook、设计 token 与规格，React `TryApp` 组件尚未落地。
  - 首轮实现聚焦 TUI selector 可验证对齐：i18n footer 文案、标记删除行危险背景、目录图标、循环导航、相对时间格式。
  - 修复阻断 `staticcheck` 的 dialog 小问题：删除未使用 modal helper，按 Lipgloss v2 用样式值派生替代 `Copy()`。
  - 用户截图反馈列表目录名几乎不可读：修复选中行前景误用 background、普通行错误套用 `SurfaceSelected`，并纠正 `COLORFGBG` 亮暗判定。
- **结论 / 决定**：
  - TUI 保持 `Enter` 选中后输出 cd 脚本的既有契约，不实现 GUI 的 FilesView。
  - selector 导航与 `try-gui/lib/use-list-navigation.ts` 对齐：`↑/↓` 与 `Ctrl-P/N` 在首尾循环。
  - 标记删除行采用 `✕` 图标、`dangerSurface` 背景、`onDanger` 前景和删除线；普通目录图标采用 `📁`。
  - 相对时间格式对齐 `try-gui/lib/try-types.ts`：`just now`、`Xm`、`Xh`、`Xd`、`Xmo`、`Xy`。
  - 列表文字使用 `foreground`；仅选中/标记行套状态背景。`COLORFGBG` 背景为 `7`/`15` 判为 light。
- **相关**：`internal/selector/`、`internal/i18n/messages.go`、`internal/dialog/`、`internal/config/config.go`、`spec/selector.md`、`spec/tui-framework.md`、`spec/i18n.md`、`spec/config.md`、`spec/architecture.md`、`README.md`

---
