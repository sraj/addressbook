import { create } from 'zustand'
import { api } from '@/lib/api'
import { AppError, showErrorToast } from '@/lib/error'
import { useBillingStore } from '@/store/billing'
import type { Bookmark } from '@/types'

interface BookmarksState {
  bookmarks: Bookmark[]
  total: number
  page: number
  size: number
  totalPages: number
  category: string
  categories: string[]
  loading: boolean
  fetchError: string | null
  quotaError: string | null
  clearQuotaError: () => void
  fetchBookmarks: (params?: { category?: string; page?: number; size?: number }) => Promise<void>
  fetchCategories: () => Promise<void>
  createBookmark: (data: Parameters<typeof api.createBookmark>[0]) => Promise<void>
  updateBookmark: (id: number, data: Parameters<typeof api.updateBookmark>[1]) => Promise<void>
  deleteBookmark: (id: number) => Promise<void>
  importBookmarks: (bookmarks: Parameters<typeof api.importBookmarks>[0]) => Promise<{ imported: number; skipped: number }>
  setCategory: (category: string) => void
  setPage: (page: number) => void
}

export const useBookmarksStore = create<BookmarksState>((set, get) => ({
  bookmarks: [],
  total: 0,
  page: 1,
  size: 20,
  totalPages: 0,
  category: '',
  categories: [],
  loading: false,
  fetchError: null,
  quotaError: null,

  clearQuotaError: () => set({ quotaError: null }),

  fetchBookmarks: async (params) => {
    set({ loading: true, fetchError: null })
    try {
      const { category, page, size } = get()
      const res = await api.getBookmarks({
        category: params?.category ?? category,
        page: params?.page ?? page,
        size: params?.size ?? size,
      })
      const planLimit = useBillingStore.getState().limits.bookmarks ?? 100
      if (planLimit > 0 && res.total >= planLimit) {
        set({ quotaError: 'bookmarks' })
      }
      set({
        bookmarks: res.data,
        total: res.total,
        page: res.page,
        size: res.size,
        totalPages: res.total_pages,
        loading: false,
      })
    } catch (err) {
      showErrorToast(err, 'Failed to load bookmarks')
      set({ fetchError: 'Failed to load bookmarks', loading: false })
    }
  },

  fetchCategories: async () => {
    try {
      const res = await api.getBookmarkCategories()
      set({ categories: res.categories })
    } catch {
      // non-critical
    }
  },

  createBookmark: async (data) => {
    try {
      await api.createBookmark(data)
      set({ quotaError: null })
      await get().fetchBookmarks({ page: 1 })
    } catch (err) {
      if (err instanceof AppError && err.status === 403 && err.message.includes('quota')) {
        set({ quotaError: 'bookmarks' })
      } else {
        showErrorToast(err, 'Failed to create bookmark')
      }
      throw err
    }
  },

  updateBookmark: async (id, data) => {
    await api.updateBookmark(id, data)
    await get().fetchBookmarks()
  },

  deleteBookmark: async (id) => {
    await api.deleteBookmark(id)
    await get().fetchBookmarks()
  },

  importBookmarks: async (items) => {
    try {
      const res = await api.importBookmarks(items)
      await get().fetchBookmarks({ page: 1 })
      await get().fetchCategories()
      return res
    } catch (err) {
      if (err instanceof AppError && err.status === 403 && err.message.includes('quota')) {
        set({ quotaError: 'bookmarks' })
      }
      throw err
    }
  },

  setCategory: (category) => {
    set({ category, page: 1 })
    get().fetchBookmarks({ category, page: 1 })
  },

  setPage: (page) => {
    set({ page })
    get().fetchBookmarks({ page })
  },
}))
