import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import AuthLayout from '@/components/layout/AuthLayout'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { toast } from '@/hooks/use-toast'
import { showErrorToast } from '@/lib/error'
import { ArrowLeft } from 'lucide-react'
import { api } from '@/lib/api'

const forgotPasswordSchema = z.object({
  email: z.string().min(1, 'Email is required').email('Invalid email address'),
})

type ForgotPasswordSchema = z.infer<typeof forgotPasswordSchema>

export default function ForgotPasswordPage() {
  const [submitted, setSubmitted] = useState(false)
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<ForgotPasswordSchema>({
    resolver: zodResolver(forgotPasswordSchema),
  })

  const onSubmit = async (data: ForgotPasswordSchema) => {
    try {
      await api.forgotPassword(data.email)
      setSubmitted(true)
      toast({ title: 'Reset link sent', description: 'Check your email for password reset instructions', variant: 'success' })
    } catch (err) {
      showErrorToast(err, 'Failed to send reset link')
    }
  }

  return (
    <AuthLayout
      title="Address Book"
      description={submitted ? 'Check your email for the reset link' : 'Enter your email to receive a reset link'}
      footer={
        <Link
          to="/login"
          className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-primary"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to sign in
        </Link>
      }
    >
      {!submitted ? (
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="email">Email</Label>
            <Input id="email" type="email" placeholder="you@example.com" {...register('email')} autoFocus />
            {errors.email && <p className="text-sm text-destructive">{errors.email.message}</p>}
          </div>
          <Button type="submit" className="w-full" disabled={isSubmitting}>
            {isSubmitting ? 'Sending…' : 'Send reset link'}
          </Button>
        </form>
      ) : (
        <div className="text-center">
          <p className="text-sm text-muted-foreground">
            If an account with that email exists, we've sent a password reset link.
          </p>
        </div>
      )}
    </AuthLayout>
  )
}
