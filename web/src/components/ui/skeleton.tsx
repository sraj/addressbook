import { forwardRef } from 'react'
import { Skeleton as NSkeleton, type SkeletonProps as NSkeletonProps } from '@mobentum/nebula-ui'
import { cn } from '@/lib/utils'

const Skeleton = forwardRef<HTMLDivElement, NSkeletonProps>(({ className, ...props }, ref) => (
  <NSkeleton ref={ref} className={cn('bg-primary/10', className)} {...props} />
))
Skeleton.displayName = 'Skeleton'

export { Skeleton }