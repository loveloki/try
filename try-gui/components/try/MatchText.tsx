'use client'

import { fuzzyMatch } from '@/lib/try-types'

interface MatchTextProps {
  text: string
  query: string
  className?: string
  matchClassName?: string
}

export function MatchText({ text, query, className, matchClassName }: MatchTextProps) {
  const segments = fuzzyMatch(text, query)
  if (!segments) {
    return <span className={className}>{text}</span>
  }
  return (
    <span className={className}>
      {segments.map((seg, i) =>
        seg.matched ? (
          <span key={i} className={matchClassName ?? 'match-char'}>
            {seg.text}
          </span>
        ) : (
          <span key={i}>{seg.text}</span>
        )
      )}
    </span>
  )
}
