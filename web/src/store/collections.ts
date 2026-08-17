import { create } from 'zustand'
import { api } from '@/lib/api'
import { showErrorToast } from '@/lib/error'
import type { Collection, Contact } from '@/types'

interface CollectionsState {
  collections: Collection[]
  loading: boolean
  fetchCollections: () => Promise<void>
  createCollection: (name: string) => Promise<Collection>
  renameCollection: (id: number, name: string) => Promise<void>
  deleteCollection: (id: number) => Promise<void>
  regenerateToken: (id: number) => Promise<Collection>
}

export const useCollectionsStore = create<CollectionsState>((set, get) => ({
  collections: [],
  loading: false,

  fetchCollections: async () => {
    set({ loading: true })
    try {
      const res = await api.getCollections()
      set({ collections: res.collections, loading: false })
    } catch (err) {
      showErrorToast(err, 'Failed to load collections')
      set({ loading: false })
    }
  },

  createCollection: async (name) => {
    const collection = await api.createCollection(name)
    await get().fetchCollections()
    return collection
  },

  renameCollection: async (id, name) => {
    await api.renameCollection(id, name)
    await get().fetchCollections()
  },

  deleteCollection: async (id) => {
    await api.deleteCollection(id)
    set({ collections: get().collections.filter((c) => c.id !== id) })
  },

  regenerateToken: async (id) => {
    const collection = await api.regenerateCollectionToken(id)
    set({
      collections: get().collections.map((c) => (c.id === id ? collection : c)),
    })
    return collection
  },
}))

export interface CollectionContacts {
  contacts: Contact[]
  total: number
  page: number
  totalPages: number
}

export function useCollectionContacts() {
  return {
    fetch: async (id: number, params: { page?: number; size?: number } = {}) => {
      const res = await api.getCollectionContacts(id, params)
      return {
        contacts: res.data,
        total: res.total,
        page: res.page,
        totalPages: res.total_pages,
      } satisfies CollectionContacts
    },
  }
}