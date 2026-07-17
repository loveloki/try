'use client'

import { SourceTab } from '@/lib/try-types'
import { cn } from '@/lib/utils'

interface SourceTabsProps {
  active: SourceTab
  counts: Record<SourceTab, number>
  onChange: (t: SourceTab) => void
}

const TABS: { id: SourceTab; label: string }[] = [
  { id: 'all', label: '全部' },
  { id: 'tries', label: 'tries' },
  { id: 'ship', label: 'ship' },
  { id: 'bug', label: 'bug' },
]

export function SourceTabs({ active, counts, onChange }: SourceTabsProps) {
  return (
    <div className="flex-none flex items-center gap-1 px-4 py-2 border-b border-[var(--try-line)] bg-background">
      {TABS.map((tab) => (
        <button
          key={tab.id}
          onClick={() => onChange(tab.id)}
          className={cn(
            'flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-mono transition-colors',
            active === tab.id
              ? 'bg-[var(--try-surface-selected)] text-[var(--try-blue)] font-semibold'
              : 'text-muted-foreground hover:text-foreground hover:bg-[var(--try-surface-hover)]',
          )}
        >
          {tab.label}
          <span
            className={cn(
              'inline-flex items-center justify-center min-w-[16px] h-4 px-1 rounded-full text-[10px] font-mono',
              active === tab.id
                ? 'bg-[var(--try-blue)]/20 text-[var(--try-blue)]'
                : 'bg-muted text-muted-foreground',
            )}
          >
            {counts[tab.id]}
          </span>
        </button>
      ))}

      {/* Keyboard hint */}
      <span className="ml-auto text-[10px] text-muted-foreground/50 font-mono hidden sm:block">
        Tab 切换
      </span>
    </div>
  )
}
