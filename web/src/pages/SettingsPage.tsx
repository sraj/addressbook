import { useLocation, useNavigate, Outlet } from 'react-router-dom'
import PageLayout from '@/components/layout/PageLayout'
import PageHeader from '@/components/layout/PageHeader'
import { Settings as SettingsIcon, CreditCard } from 'lucide-react'

export default function SettingsPage() {
  const location = useLocation()
  const navigate = useNavigate()
  const isBilling = location.pathname.endsWith('/billing')
  const showBilling = true

  return (
    <PageLayout>
      <PageHeader title="Settings" description="Manage your account and preferences" />

      <div className="flex items-center gap-0 border-b mb-6">
        <button
          type="button"
          onClick={() => navigate('/settings')}
          className={`px-4 py-2 text-sm font-medium transition-colors border-b-2 -mb-[1px] inline-flex items-center gap-2 ${
            !isBilling
              ? 'border-primary text-foreground'
              : 'border-transparent text-muted-foreground hover:text-foreground'
          }`}
        >
          <SettingsIcon className="h-4 w-4" />
          General
        </button>
        {showBilling && (
          <button
            type="button"
            onClick={() => navigate('/settings/billing')}
            className={`px-4 py-2 text-sm font-medium transition-colors border-b-2 -mb-[1px] inline-flex items-center gap-2 ${
              isBilling
                ? 'border-primary text-foreground'
                : 'border-transparent text-muted-foreground hover:text-foreground'
            }`}
          >
            <CreditCard className="h-4 w-4" />
            Billing
          </button>
        )}
      </div>

      <Outlet />
    </PageLayout>
  )
}
