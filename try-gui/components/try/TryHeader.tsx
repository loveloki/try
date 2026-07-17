'use client'

import { useEffect, useRef } from 'react'
import { Search, X, Sun, Moon } from 'lucide-react'

interface TryHeaderProps {
  title: string
  subtitle?: string
  searchQuery: string
  onSearchChange: (v: string) => void
  searchPlaceholder?: string
  theme: 'dark' | 'light'
  onThemeToggle: () => void
  showSearch?: boolean
  inputRef?: React.RefObject<HTMLInputElement | null>
}

export function TryHeader({
  title,
  subtitle,
  searchQuery,
  onSearchChange,
  searchPlaceholder = '模糊搜索…',
  theme,
  onThemeToggle,
  showSearch = true,
  inputRef: externalInputRef,
}: TryHeaderProps) {
  const internalInputRef = useRef<HTMLInputElement>(null)
  const inputRef = externalInputRef ?? internalInputRef

  // Global "/" and "Ctrl+F" shortcuts focus search
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      const isSearchShortcut = e.key === '/' || (e.ctrlKey && e.key === 'f')
      if (isSearchShortcut && document.activeElement !== inputRef.current) {
        e.preventDefault()
        inputRef.current?.focus()
      }
      if (e.key === 'Escape') {
        inputRef.current?.blur()
        onSearchChange('')
      }
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [onSearchChange, inputRef])

  return (
    <header className="flex-none border-b border-border bg-background/95 backdrop-blur-sm">
      <div className="flex items-center justify-between px-4 py-3">
        {/* Brand */}
        <div className="flex items-center gap-3">
          <div className="flex items-center justify-center w-7 h-7 rounded-md bg-[var(--try-blue-dim)] border border-[var(--try-blue)]/30">
            <span className="font-mono text-xs font-bold text-[var(--try-blue)] leading-none">try</span>
          </div>
          <div>
            <h1 className="text-sm font-semibold text-foreground leading-tight">{title}</h1>
            {subtitle && (
              <p className="text-[11px] text-muted-foreground leading-tight font-mono">{subtitle}</p>
            )}
          </div>
        </div>

        {/* Right actions */}
        <div className="flex items-center gap-2">
          <button
            onClick={onThemeToggle}
            className="flex items-center justify-center w-7 h-7 rounded-md text-muted-foreground hover:text-foreground hover:bg-[var(--try-surface-hover)] transition-colors"
            aria-label={theme === 'dark' ? '切换到亮色模式' : '切换到暗色模式'}
          >
            {theme === 'dark' ? <Sun size={14} /> : <Moon size={14} />}
          </button>
        </div>
      </div>

      {/* Search bar */}
      {showSearch && (
        <div className="px-4 pb-3">
          <div className="relative flex items-center">
            <Search size={14} className="absolute left-3 text-muted-foreground pointer-events-none" />
            <input
              ref={inputRef}
              type="text"
              value={searchQuery}
              onChange={(e) => onSearchChange(e.target.value)}
              placeholder={searchPlaceholder}
              className="w-full h-8 pl-8 pr-8 rounded-md bg-[var(--try-surface-hover)] border border-border text-sm font-mono text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-[var(--try-blue)] focus:border-[var(--try-blue)]/50 transition-all"
            />
            {searchQuery && (
              <button
                onClick={() => onSearchChange('')}
                className="absolute right-2 text-muted-foreground hover:text-foreground transition-colors"
                aria-label="清除搜索"
              >
                <X size={12} />
              </button>
            )}
            {!searchQuery && (
              <kbd className="absolute right-2 hidden sm:flex items-center justify-center h-4 px-1 rounded text-[10px] font-mono text-muted-foreground border border-border/60 bg-muted/30">
                /
              </kbd>
            )}
          </div>
        </div>
      )}
    </header>
  )
}
