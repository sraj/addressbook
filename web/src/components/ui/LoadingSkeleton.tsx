import { Skeleton } from '@/components/ui/skeleton'

interface LoadingSkeletonProps {
  count?: number
  height?: string
}

export default function LoadingSkeleton({ count = 5, height = 'h-24' }: LoadingSkeletonProps) {
  return (
    <div className="space-y-3">
      {Array.from({ length: count }).map((_, i) => (
        <Skeleton key={i} className={`${height} w-full`} />
      ))}
    </div>
  )
}
