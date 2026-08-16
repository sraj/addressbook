import { useState } from 'react'
import { cn } from '@/lib/utils'

interface CategoryFilterProps {
  categories: string[]
  selected: string
  onSelect: (category: string) => void
}

const MAX_VISIBLE = 10

function Pill({ children, active, onClick }: { children: React.ReactNode; active: boolean; onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        'rounded-full px-3 py-1.5 text-sm font-medium transition-colors shrink-0',
        active
          ? 'bg-primary text-primary-foreground'
          : 'bg-muted text-muted-foreground hover:bg-muted/80',
      )}
    >
      {children}
    </button>
  )
}

function Row({ children, active, onClick }: { children: React.ReactNode; active: boolean; onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        'rounded-md px-3 py-2 text-sm font-medium transition-colors text-left',
        active
          ? 'bg-primary text-primary-foreground'
          : 'text-muted-foreground hover:bg-muted/50',
      )}
    >
      {children}
    </button>
  )
}

export default function CategoryFilter({ categories, selected, onSelect }: CategoryFilterProps) {
  const [showAll, setShowAll] = useState(false)
  const visible = showAll ? categories : categories.slice(0, MAX_VISIBLE)
  const hiddenCount = categories.length - MAX_VISIBLE

  // Mobile: pills with overflow control
  const mobilePills = (
    <div className="flex flex-wrap gap-2 lg:hidden">
      <Pill active={selected === ''} onClick={() => onSelect('')}>All</Pill>
      {visible.map((c) => (
        <Pill key={c} active={selected === c} onClick={() => onSelect(c)}>{c}</Pill>
      ))}
      {!showAll && hiddenCount > 0 && (
        <button
          type="button"
          onClick={() => setShowAll(true)}
          className="rounded-full px-3 py-1.5 text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
        >
          +{hiddenCount}
        </button>
      )}
      {showAll && categories.length > MAX_VISIBLE && (
        <button
          type="button"
          onClick={() => setShowAll(false)}
          className="rounded-full px-3 py-1.5 text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
        >
          Show less
        </button>
      )}
    </div>
  )

  // Desktop: column layout
  const desktopColumn = (
    <div className="hidden lg:flex lg:flex-col lg:gap-1">
      <Row active={selected === ''} onClick={() => onSelect('')}>All bookmarks</Row>
      {categories.map((c) => (
        <Row key={c} active={selected === c} onClick={() => onSelect(c)}>{c}</Row>
      ))}
    </div>
  )

  return (
    <>
      {mobilePills}
      {desktopColumn}
    </>
  )
}
