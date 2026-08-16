import { useEffect, useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { useBillingStore } from '@/store/billing'
import { toast } from '@/hooks/use-toast'
import { showErrorToast } from '@/lib/error'
import { api } from '@/lib/api'
import { CreditCard, BarChart3, Receipt, ExternalLink } from 'lucide-react'

interface Invoice {
  id: string
  amount_paid: number
  currency: string
  status: string
  created: number
  period_start: number
  period_end: number
  hosted_invoice_url: string
  invoice_pdf: string
  number: string
}

interface PlanOption {
  id: number
  name: string
  price_monthly: number
  stripe_price_id: string
}

const resourceLabels: Record<string, string> = {
  contacts: 'Contacts',
  notes: 'Notes',
  bookmarks: 'Bookmarks',
}

function formatDate(ts: number) {
  return new Date(ts * 1000).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' })
}

function formatAmount(cents: number, currency: string) {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: currency.toUpperCase() }).format(cents / 100)
}

export default function BillingSettings() {
  const { usage, limits, plan, loading, fetchUsage } = useBillingStore()
  const [checkingOut, setCheckingOut] = useState(false)
  const [openingPortal, setOpeningPortal] = useState(false)
  const [invoices, setInvoices] = useState<Invoice[]>([])
  const [loadingInvoices, setLoadingInvoices] = useState(false)
  const [allPlans, setAllPlans] = useState<PlanOption[]>([])
  const [showPlanPicker, setShowPlanPicker] = useState(false)

  useEffect(() => {
    fetchUsage()
    fetchInvoices()
    fetchPlans()
  }, [])

  const fetchInvoices = async () => {
    setLoadingInvoices(true)
    try {
      const res = await api.getInvoices()
      setInvoices(res)
    } catch {
      // non-critical
    } finally {
      setLoadingInvoices(false)
    }
  }

  const fetchPlans = async () => {
    try {
      const res = await api.getPlans()
      setAllPlans(res.plans)
    } catch {
      // non-critical
    }
  }

  const handleCheckout = async (planName?: string) => {
    setCheckingOut(true)
    try {
      const res = await api.createCheckoutSession(planName)
      window.location.href = res.url
    } catch (err) {
      showErrorToast(err, 'Failed to start checkout')
    } finally {
      setCheckingOut(false)
    }
  }

  const handleChangePlan = async (planName: string) => {
    setCheckingOut(true)
    try {
      await api.changePlan(planName)
      toast({ title: `Switched to ${planName}`, variant: 'success' })
      window.location.reload()
    } catch (err) {
      showErrorToast(err, 'Failed to change plan')
    } finally {
      setCheckingOut(false)
    }
  }

  const handleManage = async () => {
    setOpeningPortal(true)
    try {
      const res = await api.createPortalSession()
      window.location.href = res.url
    } catch (err) {
      showErrorToast(err, 'Failed to open billing portal')
    } finally {
      setOpeningPortal(false)
    }
  }

  const handleCancel = async () => {
    if (!window.confirm('Are you sure you want to cancel your subscription? You will be downgraded to the Free plan.')) return
    try {
      await api.cancelSubscription()
      window.location.reload()
    } catch (err) {
      showErrorToast(err, 'Failed to cancel subscription')
    }
  }

  const currentPlanName = plan?.name || 'free'
  const isFree = currentPlanName === 'free'
  const paidPlans = allPlans.filter((p) => p.name !== 'free' && p.stripe_price_id)

  const statusBadge = (status: string) => {
    const colors: Record<string, string> = {
      paid: 'bg-success/10 text-success',
      open: 'bg-muted text-muted-foreground',
      uncollectible: 'bg-destructive/10 text-destructive',
      void: 'bg-muted text-muted-foreground',
    }
    return `inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${colors[status] || 'bg-muted text-muted-foreground'}`
  }

  return (
    <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center gap-2">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10">
              <CreditCard className="h-4 w-4 text-primary" />
            </div>
            <div>
              <CardTitle className="text-base">Current Plan</CardTitle>
              <CardDescription>You're on the {plan?.name ?? 'free'} plan</CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="text-sm text-muted-foreground">Loading…</div>
          ) : (
            <>
              <div className="text-2xl font-bold capitalize">{plan?.name ?? 'Free'}</div>
              <p className="text-sm text-muted-foreground mt-1">
                {isFree ? 'Free' : `$${((plan?.price_monthly ?? 0) / 100).toFixed(2)} / month`}
              </p>
              <div className="mt-4 flex flex-wrap gap-2">
                {isFree ? (
                  <Button size="sm" onClick={() => handleCheckout('pro')} disabled={checkingOut}>
                    {checkingOut ? 'Redirecting…' : 'Upgrade to Pro'}
                  </Button>
                ) : (
                  <>
                    <Button variant="outline" size="sm" onClick={handleManage} disabled={openingPortal}>
                      {openingPortal ? 'Loading…' : 'Manage billing'}
                    </Button>
                    <Button variant="ghost" size="sm" onClick={() => setShowPlanPicker(!showPlanPicker)}>
                      Change plan
                    </Button>
                    <button type="button" onClick={handleCancel} className="text-xs text-destructive hover:text-destructive/80 transition-colors">
                      Cancel subscription
                    </button>
                  </>
                )}
              </div>

              {showPlanPicker && paidPlans.length > 0 && (
                <div className="mt-4 space-y-2 border-t pt-4">
                  <p className="text-xs text-muted-foreground mb-2">Choose a plan:</p>
                  {paidPlans.map((p) => {
                    const isCurrent = p.name === currentPlanName
                    return (
                      <div key={p.id} className="flex items-center justify-between rounded-md border p-3">
                        <div>
                          <div className="text-sm font-medium capitalize">{p.name}</div>
                          <div className="text-xs text-muted-foreground">${(p.price_monthly / 100).toFixed(2)}/mo</div>
                        </div>
                        {isCurrent ? (
                          <span className="text-xs text-muted-foreground">Current</span>
                        ) : (
                          <Button size="sm" variant="outline" onClick={() => handleChangePlan(p.name)} disabled={checkingOut}>
                            {checkingOut ? '…' : 'Switch'}
                          </Button>
                        )}
                      </div>
                    )
                  })}
                </div>
              )}

            </>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center gap-2">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10">
              <BarChart3 className="h-4 w-4 text-primary" />
            </div>
            <div>
              <CardTitle className="text-base">Usage</CardTitle>
              <CardDescription>Resources used across your account</CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {loading ? (
            <div className="text-sm text-muted-foreground">Loading…</div>
          ) : (
            Object.entries(resourceLabels).map(([key, label]) => {
              const used = usage[key] ?? 0
              const limit = limits[key]
              if (limit === undefined) return null
              const percentage = limit === -1 ? 0 : Math.min((used / limit) * 100, 100)
              return (
                <div key={key}>
                  <div className="flex justify-between text-sm mb-1">
                    <span>{label}</span>
                    <span className="text-muted-foreground">
                      {limit === -1 ? `${used} / Unlimited` : `${used} / ${limit}`}
                    </span>
                  </div>
                  {limit !== -1 && (
                    <div className="h-2 rounded-full bg-muted overflow-hidden">
                      <div className="h-full rounded-full bg-primary transition-all" style={{ width: `${percentage}%` }} />
                    </div>
                  )}
                </div>
              )
            })
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center gap-2">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10">
              <Receipt className="h-4 w-4 text-primary" />
            </div>
            <div>
              <CardTitle className="text-base">Invoices</CardTitle>
              <CardDescription>Your billing history</CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {isFree ? (
            <p className="text-sm text-muted-foreground">No invoices yet.</p>
          ) : loadingInvoices ? (
            <div className="text-sm text-muted-foreground">Loading…</div>
          ) : invoices.length === 0 ? (
            <p className="text-sm text-muted-foreground">No invoices found.</p>
          ) : (
            <div className="space-y-2 max-h-[300px] overflow-y-auto">
              {invoices.map((inv) => (
                <div key={inv.id} className="flex items-center justify-between text-sm border-b pb-2 last:border-0">
                  <div className="min-w-0 flex-1">
                    <div className="font-medium truncate">
                      {inv.number || 'Invoice'} — {formatAmount(inv.amount_paid, inv.currency)}
                    </div>
                    <div className="text-xs text-muted-foreground">
                      {formatDate(inv.created)} • {formatDate(inv.period_start)} – {formatDate(inv.period_end)}
                    </div>
                  </div>
                  <div className="flex items-center gap-2 shrink-0 ml-2">
                    <span className={statusBadge(inv.status)}>
                      {inv.status === 'paid' ? 'Paid' : inv.status === 'open' ? 'Open' : inv.status}
                    </span>
                    {inv.hosted_invoice_url && (
                      <a
                        href={inv.hosted_invoice_url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-muted-foreground hover:text-foreground transition-colors"
                        title="View invoice"
                      >
                        <ExternalLink className="h-4 w-4" />
                      </a>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
