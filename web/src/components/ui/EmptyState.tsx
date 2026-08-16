import { type ReactNode } from 'react'
import { cn } from '@/lib/utils'

interface EmptyStateProps {
  icon: ReactNode
  heading: string
  description: string
  action?: ReactNode
  className?: string
}

export default function EmptyState({ icon, heading, description, action, className }: EmptyStateProps) {
  return (
    <div className={cn('flex flex-col items-center justify-center py-20 text-center', className)}>
      <div className="mb-4 h-12 w-12 text-muted-foreground/50">
        {icon}
      </div>
      <h3 className="text-lg font-semibold">{heading}</h3>
      <p className="mt-1 text-sm text-muted-foreground">{description}</p>
      {action && <div className="mt-4">{action}</div>}
    </div>
  )
}
