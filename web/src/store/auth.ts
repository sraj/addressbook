import { create } from 'zustand'
import { api } from '@/lib/api'
import type { User } from '@/types'

interface AuthState {
  user: User | null
  isAuthenticated: boolean
  initialized: boolean
  loading: boolean
  login: (email: string, password: string) => Promise<void>
  register: (email: string, password: string) => Promise<void>
  logout: () => Promise<void>
  checkAuth: () => Promise<void>
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isAuthenticated: false,
  initialized: false,
  loading: false,

  login: async (email, password) => {
    set({ loading: true })
    try {
      const res = await api.login(email, password)
      set({ user: res.user, isAuthenticated: true, loading: false })
    } catch (err) {
      set({ loading: false })
      throw err
    }
  },

  register: async (email, password) => {
    set({ loading: true })
    try {
      const res = await api.register(email, password)
      set({ user: res.user, isAuthenticated: true, loading: false })
    } catch (err) {
      set({ loading: false })
      throw err
    }
  },

  logout: async () => {
    set({ user: null, isAuthenticated: false, loading: false })
    try {
      await api.logout()
    } catch {
      // ignore errors — cookie already cleared locally
    }
  },

  checkAuth: async () => {
    try {
      const res = await api.me()
      set({ user: res.user, isAuthenticated: true, initialized: true })
    } catch {
      set({ user: null, isAuthenticated: false, initialized: true })
    }
  },
}))
