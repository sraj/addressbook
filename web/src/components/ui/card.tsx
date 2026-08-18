import { forwardRef } from 'react'
import {
  CardRoot as NCard,
  CardHeader as NCardHeader,
  CardTitle as NCardTitle,
  CardDescription as NCardDescription,
  CardContent as NCardContent,
  CardFooter as NCardFooter,
  type CardRootProps,
  type CardHeaderProps,
  type CardTitleProps,
  type CardDescriptionProps,
  type CardContentProps,
  type CardFooterProps,
} from '@mobentum/nebula-ui'
import { cn } from '@/lib/utils'

const Card = forwardRef<HTMLDivElement, CardRootProps>(
  ({ className, ...props }, ref) => (
    <NCard ref={ref} className={cn('bg-card text-card-foreground', className)} {...props} />
  ),
)
Card.displayName = 'Card'

const CardHeader = forwardRef<HTMLDivElement, CardHeaderProps>(
  ({ className, ...props }, ref) => (
    <NCardHeader ref={ref} className={cn('flex flex-col space-y-1.5', className)} {...props} />
  ),
)
CardHeader.displayName = 'CardHeader'

const CardTitle = forwardRef<HTMLDivElement, CardTitleProps>(
  ({ className, ...props }, ref) => (
    <NCardTitle ref={ref} className={className} {...props} />
  ),
)
CardTitle.displayName = 'CardTitle'

const CardDescription = forwardRef<HTMLDivElement, CardDescriptionProps>(
  ({ className, ...props }, ref) => (
    <NCardDescription ref={ref} className={className} {...props} />
  ),
)
CardDescription.displayName = 'CardDescription'

const CardContent = forwardRef<HTMLDivElement, CardContentProps>(
  ({ className, ...props }, ref) => (
    <NCardContent ref={ref} className={cn('pt-0', className)} {...props} />
  ),
)
CardContent.displayName = 'CardContent'

const CardFooter = forwardRef<HTMLDivElement, CardFooterProps>(
  ({ className, ...props }, ref) => (
    <NCardFooter ref={ref} className={cn('flex items-center', className)} {...props} />
  ),
)
CardFooter.displayName = 'CardFooter'

export { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter }