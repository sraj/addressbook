export interface Address {
  label: string
  line1: string
  line2?: string
  city: string
  state: string
  zip: string
  country: string
}

export interface Contact {
  id: number
  user_id: number
  collection_id?: number
  name: string
  emails: string[] | null
  phones: string[] | null
  addresses: Address[] | null
  notes: string
  created_at: string
  updated_at: string
}

export interface Collection {
  id: number
  user_id: number
  name: string
  invite_token: string
  contact_count?: number
  created_at?: string
  updated_at?: string
}

export interface LabelFormat {
  code: string
  name: string
  width: string
  height: string
  columns: number
  rows: number
  font_size_px: number
  cell_padding: string
}

export interface LabelOrder {
  id: number
  user_id: number
  collection_id?: number
  contact_count: number
  sheet_count: number
  amount_cents: number
  currency: string
  status: 'pending' | 'paid' | 'canceled'
  label_type?: string
  stripe_session_id?: string
  created_at?: string
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  size: number
  total_pages: number
}

export interface Bookmark {
  id: number
  user_id: number
  url: string
  title: string
  description?: string
  favicon_url?: string
  category?: string
  created_at: string
  updated_at: string
}

export interface Note {
  id: number
  user_id: number
  title: string
  content: string
  created_at: string
  updated_at: string
}

export enum SheetMode {
  Closed = 'closed',
  View = 'view',
  Edit = 'edit',
}

export interface UserPreferences {
  default_page: string
  page_size: number
  compact: boolean
  notify_billing: boolean
  notify_marketing: boolean
}

export interface User {
  id: number
  email: string
  name: string
  role?: string
  preferences?: UserPreferences
}

export interface Plan {
  id: number
  name: string
  price_monthly: number
  limits?: Record<string, number>
}

export interface UsageResponse {
  usage: Record<string, number>
  plan: Plan
  limits: Record<string, number>
}

export interface ApiError {
  error: string
  fields?: Array<{
    field: string
    tag: string
    value?: unknown
    message?: string
  }>
}
