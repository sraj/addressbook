import {
  TabsRoot as NTabs,
  TabsList as NTabsList,
  TabsTrigger as NTabsTrigger,
  TabsContent as NTabsContent,
  type TabsRootProps,
  type TabsListProps,
  type TabsTabProps,
  type TabsPanelProps,
} from '@mobentum/nebula-ui'
import { cn } from '@/lib/utils'

export function Tabs({ className, ...props }: TabsRootProps) {
  return <NTabs className={className} {...props} />
}

export function TabsList({ className, ...props }: TabsListProps) {
  return (
    <NTabsList
      className={cn('inline-flex h-9 items-center justify-center rounded-lg bg-muted p-1 text-muted-foreground', className)}
      {...props}
    />
  )
}

export function TabsTrigger({ className, ...props }: TabsTabProps) {
  return (
    <NTabsTrigger
      className={cn(
        'inline-flex items-center justify-center whitespace-nowrap rounded-md px-3 py-1 text-sm font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 aria-selected:bg-background aria-selected:text-foreground aria-selected:shadow',
        className,
      )}
      {...props}
    />
  )
}

export function TabsContent({ className, ...props }: TabsPanelProps) {
  return <NTabsContent className={cn('mt-4', className)} {...props} />
}