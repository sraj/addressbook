import { createContext, useContext, useState, type ReactNode } from 'react'
import { cn } from '@/lib/utils'

interface TabsContextValue {
  value: string
  setValue: (v: string) => void
}

const TabsContext = createContext<TabsContextValue>({ value: '', setValue: () => {} })

interface TabsProps {
  defaultValue?: string
  value?: string
  onValueChange?: (value: string) => void
  children: ReactNode
}

export function Tabs({ defaultValue = '', value, onValueChange, children }: TabsProps) {
  const [internal, setInternal] = useState(defaultValue)
  const current = value ?? internal
  const setValue = (v: string) => {
    setInternal(v)
    onValueChange?.(v)
  }
  return (
    <TabsContext.Provider value={{ value: current, setValue }}>
      <div>{children}</div>
    </TabsContext.Provider>
  )
}

interface TabsListProps {
  children: ReactNode
  className?: string
}

export function TabsList({ children, className }: TabsListProps) {
  return (
    <div
      className={cn(
        'inline-flex h-9 items-center justify-center rounded-lg bg-muted p-1 text-muted-foreground',
        className,
      )}
    >
      {children}
    </div>
  )
}

interface TabsTriggerProps {
  value: string
  children: ReactNode
  className?: string
}

export function TabsTrigger({ value, children, className }: TabsTriggerProps) {
  const { value: current, setValue } = useContext(TabsContext)
  const active = current === value
  return (
    <button
      type="button"
      role="tab"
      aria-selected={active}
      onClick={() => setValue(value)}
      className={cn(
        'inline-flex items-center justify-center whitespace-nowrap rounded-md px-3 py-1 text-sm font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50',
        active && 'bg-background text-foreground shadow',
        className,
      )}
    >
      {children}
    </button>
  )
}

interface TabsContentProps {
  value: string
  children: ReactNode
  className?: string
}

export function TabsContent({ value, children, className }: TabsContentProps) {
  const { value: current } = useContext(TabsContext)
  if (current !== value) return null
  return <div className={cn('mt-4', className)}>{children}</div>
}