import { useEffect, useState } from 'react'
import { api } from '@/lib/api'
import { showErrorToast } from '@/lib/error'
import PageLayout from '@/components/layout/PageLayout'
import PageHeader from '@/components/layout/PageHeader'
import LoadingSkeleton from '@/components/ui/LoadingSkeleton'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { toast } from '@/hooks/use-toast'
import { cn } from '@/lib/utils'
import { Users, CreditCard, RefreshCw } from 'lucide-react'

interface UserRow {
  id: number
  email: string
  role: string
  status: string
  created_at: string
  plan_name: string
  subscription_status: string
  subscription_end: string
}

function formatDate(ts: string) {
  if (!ts) return ''
  const d = new Date(ts)
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
}

interface PlanRow {
  id: number
  name: string
  price_monthly: number
  stripe_price_id: string
}

export default function AdminPage() {
  const [tab, setTab] = useState<'users' | 'plans'>('users')
  const [users, setUsers] = useState<UserRow[]>([])
  const [plans, setPlans] = useState<PlanRow[]>([])
  const [loading, setLoading] = useState(true)
  const [syncing, setSyncing] = useState(false)
  const [editingPrice, setEditingPrice] = useState<Record<number, string>>({})

  const fetchUsers = async () => {
    try {
      const res = await api.getAdminUsers()
      setUsers(res.users)
    } catch (err) {
      showErrorToast(err, 'Failed to load users')
    }
  }

  const fetchPlans = async () => {
    try {
      const res = await api.getAdminPlans()
      setPlans(res.plans)
      const prices: Record<number, string> = {}
      for (const p of res.plans) {
        prices[p.id] = p.stripe_price_id || ''
      }
      setEditingPrice(prices)
    } catch (err) {
      showErrorToast(err, 'Failed to load plans')
    }
  }

  useEffect(() => {
    setLoading(true)
    Promise.all([fetchUsers(), fetchPlans()]).finally(() => setLoading(false))
  }, [])

  const toggleStatus = async (user: UserRow) => {
    const newStatus = user.status === 'active' ? 'suspended' : 'active'
    try {
      await api.updateUserStatus(user.id, newStatus)
      toast({ title: `User ${newStatus}`, variant: 'success' })
      fetchUsers()
    } catch (err) {
      showErrorToast(err, 'Failed to update user')
    }
  }

  const savePriceID = async (planId: number) => {
    const priceID = editingPrice[planId] || ''
    try {
      await api.updatePlanPriceID(planId, priceID)
      toast({ title: 'Price ID updated', variant: 'success' })
      fetchPlans()
    } catch (err) {
      showErrorToast(err, 'Failed to update price ID')
    }
  }

  const handleSync = async () => {
    setSyncing(true)
    try {
      await api.syncPlansFromStripe()
      toast({ title: 'Prices synced from Stripe', variant: 'success' })
      fetchPlans()
    } catch (err) {
      showErrorToast(err, 'Failed to sync prices')
    } finally {
      setSyncing(false)
    }
  }

  return (
    <PageLayout>
      <PageHeader title="Admin" description="Manage users and accounts" />

      <div className="flex items-center gap-0 border-b mb-6">
        <button
          type="button"
          onClick={() => setTab('users')}
          className={cn(
            'px-4 py-2 text-sm font-medium transition-colors border-b-2 -mb-[1px] inline-flex items-center gap-2',
            tab === 'users' ? 'border-primary text-foreground' : 'border-transparent text-muted-foreground hover:text-foreground',
          )}
        >
          <Users className="h-4 w-4" />
          Users
        </button>
        <button
          type="button"
          onClick={() => setTab('plans')}
          className={cn(
            'px-4 py-2 text-sm font-medium transition-colors border-b-2 -mb-[1px] inline-flex items-center gap-2',
            tab === 'plans' ? 'border-primary text-foreground' : 'border-transparent text-muted-foreground hover:text-foreground',
          )}
        >
          <CreditCard className="h-4 w-4" />
          Plans
        </button>
      </div>

      {tab === 'users' && (
        loading ? (
          <LoadingSkeleton height="h-12" />
        ) : (
          <div className="rounded-lg border bg-background overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b">
                  <th className="text-left p-3 font-medium">Email</th>
                  <th className="text-left p-3 font-medium">Role</th>
                  <th className="text-left p-3 font-medium">Plan</th>
                  <th className="text-left p-3 font-medium">Status</th>
                  <th className="text-left p-3 font-medium">Subscription</th>
                  <th className="text-left p-3 font-medium">Renewal</th>
                  <th className="text-right p-3 font-medium">Actions</th>
                </tr>
              </thead>
              <tbody>
                {users.map((user) => (
                  <tr key={user.id} className="border-b last:border-0">
                    <td className="p-3">{user.email}</td>
                    <td className="p-3 capitalize text-sm">{user.role}</td>
                    <td className="p-3">
                      <span className="capitalize text-sm">{user.plan_name || '-'}</span>
                    </td>
                    <td className="p-3">
                      <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                        user.status === 'active' ? 'bg-success/10 text-success' : 'bg-destructive/10 text-destructive'
                      }`}>
                        {user.status}
                      </span>
                    </td>
                    <td className="p-3 text-xs text-muted-foreground">{user.subscription_status || '-'}</td>
                    <td className="p-3 text-xs text-muted-foreground">{formatDate(user.subscription_end)}</td>
                    <td className="p-3 text-right">
                      {user.role !== 'admin' && (
                        <Button variant="outline" size="sm" onClick={() => toggleStatus(user)}>
                          {user.status === 'active' ? 'Suspend' : 'Activate'}
                        </Button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )
      )}

      {tab === 'plans' && (
        <div className="space-y-4">
          <div className="flex justify-end">
            <Button variant="outline" size="sm" className="gap-2" onClick={handleSync} disabled={syncing}>
              <RefreshCw className={`h-4 w-4 ${syncing ? 'animate-spin' : ''}`} />
              Sync from Stripe
            </Button>
          </div>

          {loading ? (
            <LoadingSkeleton height="h-12" />
          ) : (
            <div className="rounded-lg border bg-background overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b">
                    <th className="text-left p-3 font-medium">Plan</th>
                    <th className="text-left p-3 font-medium">Price</th>
                    <th className="text-left p-3 font-medium">Stripe Price ID</th>
                    <th className="text-right p-3 font-medium">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {plans.map((plan) => (
                    <tr key={plan.id} className="border-b last:border-0">
                      <td className="p-3 font-medium capitalize">{plan.name}</td>
                      <td className="p-3">${(plan.price_monthly / 100).toFixed(2)}/mo</td>
                      <td className="p-3">
                        <Input
                          value={editingPrice[plan.id] ?? ''}
                          onChange={(e) => setEditingPrice((p) => ({ ...p, [plan.id]: e.target.value }))}
                          placeholder="price_xxx"
                          className="h-8 text-xs font-mono max-w-[280px]"
                        />
                      </td>
                      <td className="p-3 text-right">
                        <Button variant="outline" size="sm" onClick={() => savePriceID(plan.id)}>
                          Save
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}
    </PageLayout>
  )
}
