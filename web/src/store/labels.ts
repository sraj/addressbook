import { create } from 'zustand'
import { api } from '@/lib/api'
import { showErrorToast } from '@/lib/error'
import type { LabelFormat, LabelOrder } from '@/types'

interface LabelsState {
  orders: LabelOrder[]
  formats: LabelFormat[]
  loading: boolean
  fetchOrders: () => Promise<void>
  fetchFormats: () => Promise<void>
  createOrder: (collectionId?: number, email?: string, format?: string) => Promise<{ order: LabelOrder; url: string }>
}

export const useLabelsStore = create<LabelsState>((set, get) => ({
  orders: [],
  formats: [],
  loading: false,

  fetchOrders: async () => {
    set({ loading: true })
    try {
      const res = await api.getLabelOrders()
      set({ orders: res.orders, loading: false })
    } catch (err) {
      showErrorToast(err, 'Failed to load label orders')
      set({ loading: false })
    }
  },

  fetchFormats: async () => {
    try {
      const res = await api.getLabelFormats()
      set({ formats: res.formats })
    } catch (err) {
      showErrorToast(err, 'Failed to load label formats')
    }
  },

  createOrder: async (collectionId, email, format) => {
    const res = await api.createLabelOrder(collectionId, email, format)
    await get().fetchOrders()
    return res
  },
}))