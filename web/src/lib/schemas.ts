import { z } from 'zod'

export const loginSchema = z.object({
  email: z.string().min(1, 'Email is required').email('Invalid email address'),
  password: z.string().min(1, 'Password is required').min(8, 'Password must be at least 8 characters'),
})

export type LoginSchema = z.infer<typeof loginSchema>

export const registerSchema = z
  .object({
    email: z.string().min(1, 'Email is required').email('Invalid email address'),
    password: z.string().min(1, 'Password is required').min(8, 'Password must be at least 8 characters'),
    confirmPassword: z.string().min(1, 'Please confirm your password'),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: 'Passwords do not match',
    path: ['confirmPassword'],
  })

export type RegisterSchema = z.infer<typeof registerSchema>

export const addressSchema = z.object({
  label: z.enum(['Home', 'Office', 'Other'], { required_error: 'Label is required' }),
  line1: z.string().min(1, 'Street address is required').max(200, 'Street address must be at most 200 characters'),
  line2: z.string().max(200, 'Line 2 must be at most 200 characters').optional().or(z.literal('')),
  city: z.string().min(1, 'City is required').max(100, 'City must be at most 100 characters'),
  state: z.string().min(1, 'State is required').max(100, 'State must be at most 100 characters'),
  zip: z.string().min(1, 'ZIP code is required').max(20, 'ZIP must be at most 20 characters'),
  country: z.string().min(1, 'Country is required').max(100, 'Country must be at most 100 characters'),
})

export const contactSchema = z.object({
  name: z.string().min(1, 'Name is required').max(200, 'Name must be at most 200 characters'),
  emails: z
    .array(z.string().email('Invalid email address'))
    .min(1, 'At least one email is required'),
  phones: z
    .array(z.string().min(1, 'Phone is required').max(30, 'Phone must be at most 30 characters'))
    .min(1, 'At least one phone number is required'),
  addresses: z.array(addressSchema).min(1, 'At least one address is required'),
  notes: z.string().max(2000, 'Notes must be at most 2000 characters').optional().or(z.literal('')),
})

export type ContactSchema = z.infer<typeof contactSchema>
