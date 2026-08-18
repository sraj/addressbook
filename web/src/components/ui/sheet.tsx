import { forwardRef, type HTMLAttributes } from 'react'
import {
  DrawerRoot as NDrawerRoot,
  DrawerContent as NDrawerContent,
  DrawerTitle as NDrawerTitle,
  DrawerDescription as NDrawerDescription,
  DrawerClose as NDrawerClose,
  type DrawerTitleProps,
  type DrawerDescriptionProps,
} from '@mobentum/nebula-ui'
import { X } from 'lucide-react'
import { cn } from '@/lib/utils'

const Sheet = NDrawerRoot

const SheetContent = forwardRef<HTMLDivElement, React.ComponentProps<typeof NDrawerContent> & { side?: 'left' | 'right' }>(
  ({ className, children, side = 'right', ...props }, ref) => (
    <NDrawerContent
      ref={ref}
      className={cn(
        'max-w-full',
        side === 'left' && 'end-auto inset-y-0 left-0 border-s-0 border-r border-nb-border',
        className,
      )}
      {...props}
    >
      {children}
      <NDrawerClose className="absolute right-4 top-4 rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:pointer-events-none">
        <X className="h-4 w-4" />
        <span className="sr-only">Close</span>
      </NDrawerClose>
    </NDrawerContent>
  ),
)
SheetContent.displayName = 'SheetContent'

const SheetHeader = ({ className, ...props }: HTMLAttributes<HTMLDivElement>) => (
  <div className={cn('flex flex-col space-y-1.5', className)} {...props} />
)
SheetHeader.displayName = 'SheetHeader'

const SheetFooter = ({ className, ...props }: HTMLAttributes<HTMLDivElement>) => (
  <div className={cn('flex flex-col-reverse sm:flex-row sm:justify-end sm:space-x-2', className)} {...props} />
)
SheetFooter.displayName = 'SheetFooter'

const SheetTitle = forwardRef<HTMLHeadingElement, DrawerTitleProps>(
  ({ className, ...props }, ref) => (
    <NDrawerTitle ref={ref} className={cn('text-lg font-semibold text-foreground', className)} {...props} />
  ),
)
SheetTitle.displayName = 'SheetTitle'

const SheetDescription = forwardRef<HTMLDivElement, DrawerDescriptionProps>(
  ({ className, ...props }, ref) => (
    <NDrawerDescription ref={ref} className={cn('text-sm text-muted-foreground', className)} {...props} />
  ),
)
SheetDescription.displayName = 'SheetDescription'

export { Sheet, SheetContent, SheetHeader, SheetFooter, SheetTitle, SheetDescription }