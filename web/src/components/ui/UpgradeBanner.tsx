import { Button } from '@/components/ui/button'
import { useNavigate } from 'react-router-dom'

interface UpgradeBannerProps {
  resource: string
}

export default function UpgradeBanner({ resource }: UpgradeBannerProps) {
  const navigate = useNavigate()

  return (
    <div className="rounded-lg border border-amber-200 bg-amber-50 dark:border-amber-800 dark:bg-amber-950/20 p-4 mb-4">
      <div className="flex items-center justify-between gap-4">
        <div>
          <p className="text-sm font-medium text-amber-800 dark:text-amber-300">
            {resource} limit reached
          </p>
          <p className="text-xs text-amber-700 dark:text-amber-400 mt-0.5">
            You've hit the free plan limit for {resource.toLowerCase()}. Upgrade to Pro for unlimited access.
          </p>
        </div>
        <Button
          size="sm"
          className="shrink-0 bg-amber-600 hover:bg-amber-700 text-white"
          onClick={() => navigate('/settings/billing')}
        >
          Upgrade
        </Button>
      </div>
    </div>
  )
}
