import { create } from 'zustand'
import { api } from '@/lib/api'
import { AppError, showErrorToast } from '@/lib/error'
import { useBillingStore } from '@/store/billing'
import type { Contact } from '@/types'

interface ContactsState {
  contacts: Contact[]
  total: number
  page: number
  size: number
  totalPages: number
  searchQuery: string
  loading: boolean
  fetchError: string | null
  quotaError: string | null
  clearQuotaError: () => void
  fetchContacts: (params?: { q?: string; page?: number; size?: number }) => Promise<void>
  createContact: (data: Parameters<typeof api.createContact>[0]) => Promise<void>
  updateContact: (id: number, data: Parameters<typeof api.updateContact>[1]) => Promise<void>
  deleteContact: (id: number) => Promise<void>
  setSearch: (q: string) => void
  setPage: (page: number) => void
}

export const useContactsStore = create<ContactsState>((set, get) => ({
  contacts: [],
  total: 0,
  page: 1,
  size: 20,
  totalPages: 0,
  searchQuery: '',
  loading: false,
  fetchError: null,
  quotaError: null,

  clearQuotaError: () => set({ quotaError: null }),

  fetchContacts: async (params) => {
    set({ loading: true, fetchError: null })
    try {
      const { searchQuery, page, size } = get()
      const res = await api.getContacts({
        q: params?.q ?? searchQuery,
        page: params?.page ?? page,
        size: params?.size ?? size,
      })
      const planLimit = useBillingStore.getState().limits.contacts ?? 50
      if (planLimit > 0 && res.total >= planLimit) {
        set({ quotaError: 'contacts' })
      }
      set({
        contacts: res.data,
        total: res.total,
        page: res.page,
        size: res.size,
        totalPages: res.total_pages,
        loading: false,
      })
    } catch (err) {
      showErrorToast(err, 'Failed to load contacts')
      set({ fetchError: 'Failed to load contacts', loading: false })
    }
  },

  createContact: async (data) => {
    try {
      await api.createContact(data)
      set({ quotaError: null })
      await get().fetchContacts({ page: 1 })
    } catch (err) {
      if (err instanceof AppError && err.status === 403 && err.message.includes('quota')) {
        set({ quotaError: 'contacts' })
      } else {
        showErrorToast(err, 'Failed to create contact')
      }
      throw err
    }
  },

  updateContact: async (id, data) => {
    await api.updateContact(id, data)
    await get().fetchContacts()
  },

  deleteContact: async (id) => {
    await api.deleteContact(id)
    await get().fetchContacts()
  },

  setSearch: (q) => {
    set({ searchQuery: q, page: 1 })
    get().fetchContacts({ q, page: 1 })
  },

  setPage: (page) => {
    set({ page })
    get().fetchContacts({ page })
  },
}))
