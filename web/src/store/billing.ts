import { create } from 'zustand'
import { api } from '@/lib/api'
import { showErrorToast } from '@/lib/error'
import type { Plan } from '@/types'

interface BillingState {
  usage: Record<string, number>
  limits: Record<string, number>
  plan: Plan | null
  loading: boolean
  fetchUsage: () => Promise<void>
}

export const useBillingStore = create<BillingState>((set) => ({
  usage: {},
  limits: {},
  plan: null,
  loading: false,

  fetchUsage: async () => {
    set({ loading: true })
    try {
      const res = await api.getBillingUsage()
      set({ usage: res.usage, limits: res.limits, plan: res.plan, loading: false })
    } catch (err) {
      showErrorToast(err, 'Failed to load billing info')
      set({ loading: false })
    }
  },
}))
