import { create } from 'zustand'
import { api } from '@/lib/api'
import { AppError, showErrorToast } from '@/lib/error'
import { useBillingStore } from '@/store/billing'
import type { Note } from '@/types'

interface NotesState {
  notes: Note[]
  total: number
  page: number
  size: number
  totalPages: number
  loading: boolean
  fetchError: string | null
  quotaError: string | null
  clearQuotaError: () => void
  fetchNotes: (params?: { page?: number; size?: number }) => Promise<void>
  createNote: (data: { title: string; content: string }) => Promise<void>
  updateNote: (id: number, data: { title: string; content: string }) => Promise<void>
  deleteNote: (id: number) => Promise<void>
  setPage: (page: number) => void
}

export const useNotesStore = create<NotesState>((set, get) => ({
  notes: [],
  total: 0,
  page: 1,
  size: 20,
  totalPages: 0,
  loading: false,
  fetchError: null,
  quotaError: null,

  clearQuotaError: () => set({ quotaError: null }),

  fetchNotes: async (params) => {
    set({ loading: true, fetchError: null })
    try {
      const { page, size } = get()
      const res = await api.getNotes({
        page: params?.page ?? page,
        size: params?.size ?? size,
      })
      const planLimit = useBillingStore.getState().limits.notes ?? 50
      if (planLimit > 0 && res.total >= planLimit) {
        set({ quotaError: 'notes' })
      }
      set({
        notes: res.data,
        total: res.total,
        page: res.page,
        size: res.size,
        totalPages: res.total_pages,
        loading: false,
      })
    } catch (err) {
      showErrorToast(err, 'Failed to load notes')
      set({ fetchError: 'Failed to load notes', loading: false })
    }
  },

  createNote: async (data) => {
    try {
      await api.createNote(data)
      set({ quotaError: null })
      await get().fetchNotes({ page: 1 })
    } catch (err) {
      if (err instanceof AppError && err.status === 403 && err.message.includes('quota')) {
        set({ quotaError: 'notes' })
      } else {
        showErrorToast(err, 'Failed to create note')
      }
      throw err
    }
  },

  updateNote: async (id, data) => {
    await api.updateNote(id, data)
    await get().fetchNotes()
  },

  deleteNote: async (id) => {
    await api.deleteNote(id)
    await get().fetchNotes()
  },

  setPage: (page) => {
    set({ page })
    get().fetchNotes({ page })
  },
}))
