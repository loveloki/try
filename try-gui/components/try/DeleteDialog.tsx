'use client'

import { TryEntry } from '@/lib/try-types'
import { AlertTriangle } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useState } from 'react'

interface DeleteDialogProps {
  entries: TryEntry[]
  onConfirm: () => void
  onCancel: () => void
}

export function DeleteDialog({ entries, onConfirm, onCancel }: DeleteDialogProps) {
  const [focused, setFocused] = useState<'yes' | 'no'>('no')

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div
        className="w-80 rounded-xl border border-[var(--try-danger)]/30 bg-card shadow-2xl"
        role="dialog"
        aria-modal="true"
        aria-labelledby="delete-dialog-title"
      >
        {/* Header */}
        <div className="flex items-center gap-2.5 px-5 py-4 border-b border-[var(--try-danger)]/20">
          <AlertTriangle size={16} className="text-[var(--try-danger)] flex-none" />
          <h2 id="delete-dialog-title" className="text-sm font-semibold text-[var(--try-danger)]">
            确认删除 {entries.length} 个目录
          </h2>
        </div>

        {/* Entry list */}
        <div className="px-5 py-3 max-h-48 overflow-y-auto">
          {entries.map((e) => (
            <div
              key={e.id}
              className="flex items-center gap-2 py-1.5 border-b border-[var(--try-line)] last:border-0"
            >
              <span className="text-[var(--try-danger)]/60 text-xs">✕</span>
              <span className="font-mono text-xs text-[var(--try-danger)] line-through opacity-80">
                {e.name}
              </span>
            </div>
          ))}
        </div>

        {/* Actions */}
        <div className="flex items-center gap-2 px-5 py-4 border-t border-[var(--try-line)]">
          <button
            onMouseEnter={() => setFocused('no')}
            onFocus={() => setFocused('no')}
            onClick={onCancel}
            className={cn(
              'flex-1 h-8 rounded-md text-xs font-mono transition-all border',
              focused === 'no'
                ? 'bg-[var(--try-surface-selected)] border-[var(--try-blue)]/40 text-foreground'
                : 'border-border text-muted-foreground hover:text-foreground hover:bg-[var(--try-surface-hover)]',
            )}
          >
            取消（Esc）
          </button>
          <button
            onMouseEnter={() => setFocused('yes')}
            onFocus={() => setFocused('yes')}
            onClick={onConfirm}
            className={cn(
              'flex-1 h-8 rounded-md text-xs font-mono font-semibold transition-all border',
              focused === 'yes'
                ? 'bg-[var(--try-danger)] border-[var(--try-danger)] text-white'
                : 'border-[var(--try-danger)]/40 text-[var(--try-danger)] hover:bg-[var(--try-danger)]/10',
            )}
          >
            确认删除（Y）
          </button>
        </div>
      </div>
    </div>
  )
}
