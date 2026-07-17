'use client'

import { useState, useEffect } from 'react'
import { TryEntry } from '@/lib/try-types'
import { SelectorView } from './SelectorView'
import { FilesView } from './FilesView'

export type AppView = 'selector' | 'files'

export function TryApp() {
  const [theme, setTheme] = useState<'dark' | 'light'>('dark')

  // Apply dark class on mount
  useEffect(() => {
    document.documentElement.classList.add('dark')
    document.documentElement.classList.remove('light')
  }, [])
  const [view, setView] = useState<AppView>('selector')
  const [selectedEntry, setSelectedEntry] = useState<TryEntry | null>(null)

  const handleThemeToggle = () => {
    const next = theme === 'dark' ? 'light' : 'dark'
    setTheme(next)
    // Apply class to html element for CSS targeting
    document.documentElement.classList.toggle('dark', next === 'dark')
    document.documentElement.classList.toggle('light', next === 'light')
  }

  const handleOpen = (entry: TryEntry) => {
    setSelectedEntry(entry)
    setView('files')
  }

  const handleBack = () => {
    setView('selector')
  }

  return (
    <div
      className="w-full h-screen flex items-center justify-center p-4 bg-[var(--try-surface-hover)]"
      style={{ background: theme === 'dark' ? '#0d0e14' : '#e8e8e8' }}
    >
      {/* Window chrome */}
      <div
        className="relative w-full max-w-2xl h-[600px] rounded-xl overflow-hidden shadow-2xl border border-border flex flex-col"
        style={{
          background: theme === 'dark'
            ? 'oklch(0.13 0.005 264)'
            : 'oklch(0.98 0 0)',
          boxShadow: theme === 'dark'
            ? '0 32px 64px rgba(0,0,0,0.7), 0 0 0 1px rgba(255,255,255,0.05)'
            : '0 32px 64px rgba(0,0,0,0.2), 0 0 0 1px rgba(0,0,0,0.08)',
        }}
      >
        {/* macOS-style traffic lights */}
        <div className="absolute top-3.5 left-4 z-20 flex items-center gap-1.5">
          <div className="w-3 h-3 rounded-full bg-[#FF5F57] hover:brightness-110 cursor-default transition-all" />
          <div className="w-3 h-3 rounded-full bg-[#FFBD2E] hover:brightness-110 cursor-default transition-all" />
          <div className="w-3 h-3 rounded-full bg-[#28C840] hover:brightness-110 cursor-default transition-all" />
        </div>

        {/* Main content */}
        <div className="flex flex-col flex-1 overflow-hidden mt-0">
          {view === 'selector' ? (
            <SelectorView
              theme={theme}
              onThemeToggle={handleThemeToggle}
              onOpen={handleOpen}
            />
          ) : selectedEntry ? (
            <FilesView
              entry={selectedEntry}
              theme={theme}
              onThemeToggle={handleThemeToggle}
              onBack={handleBack}
            />
          ) : null}
        </div>
      </div>

      {/* Keyboard shortcut hint below window */}
      <div className="absolute bottom-4 left-1/2 -translate-x-1/2 flex items-center gap-4 text-[11px] font-mono opacity-40"
        style={{ color: theme === 'dark' ? '#fff' : '#000' }}>
        <span>双击目录进入文件视图</span>
        <span>·</span>
        <span>⌃T 创建</span>
        <span>·</span>
        <span>⌃D 删除模式</span>
        <span>·</span>
        <span>/ 搜索</span>
      </div>
    </div>
  )
}
