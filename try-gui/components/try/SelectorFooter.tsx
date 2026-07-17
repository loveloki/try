'use client'

import { cn } from '@/lib/utils'

interface SelectorFooterProps {
  totalCount: number
  filteredCount: number
  deleteMode: boolean
  markedCount: number
  showCreateInput: boolean
  createName: string
  onCreateNameChange: (v: string) => void
  onCreateSubmit: () => void
  onDeleteModeToggle: () => void
  onDeleteConfirm: () => void
}

export function SelectorFooter({
  totalCount,
  filteredCount,
  deleteMode,
  markedCount,
  showCreateInput,
  createName,
  onCreateNameChange,
  onCreateSubmit,
  onDeleteModeToggle,
  onDeleteConfirm,
}: SelectorFooterProps) {
  return (
    <footer className="flex-none border-t border-border bg-background">
      {/* Create input */}
      {showCreateInput && (
        <div className="px-4 py-2 border-b border-[var(--try-line)]">
          <div className="flex items-center gap-2">
            <span className="text-[var(--try-accent)] text-xs font-mono">+ 新建</span>
            <input
              autoFocus
              type="text"
              value={createName}
              onChange={(e) => onCreateNameChange(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' && !e.nativeEvent.isComposing) onCreateSubmit()
              }}
              placeholder="目录名称（自动添加日期后缀）"
              className="flex-1 h-7 px-2 rounded bg-[var(--try-surface-hover)] border border-[var(--try-accent)]/40 text-xs font-mono text-foreground placeholder:text-muted-foreground/60 focus:outline-none focus:ring-1 focus:ring-[var(--try-accent)]/50"
            />
            <kbd className="text-[10px] font-mono text-muted-foreground border border-border/60 rounded px-1 py-0.5">
              Enter
            </kbd>
          </div>
        </div>
      )}

      {/* Status bar */}
      <div className="flex items-center justify-between px-4 py-2">
        {/* Left: counts */}
        <div className="flex items-center gap-3">
          {deleteMode ? (
            <span className="text-xs font-mono font-semibold text-[var(--try-danger)] tracking-wide">
              DELETE MODE
            </span>
          ) : (
            <span className="text-[11px] font-mono text-muted-foreground">
              {filteredCount === totalCount
                ? `${totalCount} 项`
                : `${filteredCount} / ${totalCount} 项`}
            </span>
          )}
          {deleteMode && markedCount > 0 && (
            <span className="text-[11px] font-mono text-[var(--try-danger)]/80">
              已标记 {markedCount} 项
            </span>
          )}
        </div>

        {/* Right: shortcuts */}
        <div className="flex items-center gap-2">
          {deleteMode ? (
            <>
              <button
                onClick={onDeleteModeToggle}
                className="text-[11px] font-mono text-muted-foreground hover:text-foreground transition-colors"
              >
                <kbd className="border border-border/60 rounded px-1 py-0.5">Esc</kbd> 取消
              </button>
              {markedCount > 0 && (
                <button
                  onClick={onDeleteConfirm}
                  className="text-[11px] font-mono text-[var(--try-danger)] hover:text-[var(--try-danger)]/80 transition-colors"
                >
                  <kbd className="border border-[var(--try-danger)]/40 rounded px-1 py-0.5">Enter</kbd> 确认删除
                </button>
              )}
            </>
          ) : (
            <div className="hidden sm:flex items-center gap-3 text-[10px] font-mono text-muted-foreground/60">
              <span>
                <kbd className="border border-border/40 rounded px-1">↑↓</kbd> 导航
              </span>
              <span>
                <kbd className="border border-border/40 rounded px-1">Enter</kbd> 打开
              </span>
              <span>
                <kbd className="border border-border/40 rounded px-1">⌃T</kbd> 新建
              </span>
              <span>
                <kbd className="border border-border/40 rounded px-1">⌃D</kbd> 删除
              </span>
            </div>
          )}
        </div>
      </div>
    </footer>
  )
}
