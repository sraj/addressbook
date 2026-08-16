import { forwardRef, type ComponentRef, type ComponentPropsWithoutRef, type HTMLAttributes } from 'react'
import {
  Root as SheetRoot,
  Trigger as SheetTriggerPrimitive,
  Close as SheetClosePrimitive,
  Portal as SheetPortal,
  Overlay as SheetOverlayPrimitive,
  Content as SheetContentPrimitive,
  Title as SheetTitlePrimitive,
  Description as SheetDescriptionPrimitive,
} from '@radix-ui/react-dialog'
import { X } from 'lucide-react'
import { cn } from '@/lib/utils'

const Sheet = SheetRoot
const SheetTrigger = SheetTriggerPrimitive
const SheetClose = SheetClosePrimitive

const SheetOverlay = forwardRef<
  ComponentRef<typeof SheetOverlayPrimitive>,
  ComponentPropsWithoutRef<typeof SheetOverlayPrimitive>
>(({ className, ...props }, ref) => (
  <SheetOverlayPrimitive
    ref={ref}
    className={cn(
      'fixed inset-0 z-50 bg-black/80 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0',
      className,
    )}
    {...props}
  />
))
SheetOverlay.displayName = 'SheetOverlay'

const SheetContent = forwardRef<
  ComponentRef<typeof SheetContentPrimitive>,
  ComponentPropsWithoutRef<typeof SheetContentPrimitive> & { side?: 'left' | 'right' }
>(({ className, children, side = 'right', ...props }, ref) => (
  <SheetPortal>
    <SheetOverlay />
    <SheetContentPrimitive
      ref={ref}
      className={cn(
        'fixed z-50 gap-4 bg-background p-6 shadow-lg transition ease-in-out data-[state=open]:animate-in data-[state=closed]:animate-out',
        side === 'right' &&
          'inset-y-0 right-0 h-full w-full sm:max-w-lg border-l data-[state=closed]:slide-out-to-right data-[state=open]:slide-in-from-right',
        side === 'left' &&
          'inset-y-0 left-0 h-full w-full sm:max-w-lg border-r data-[state=closed]:slide-out-to-left data-[state=open]:slide-in-from-left',
        className,
      )}
      {...props}
    >
      {children}
      <SheetClosePrimitive className="absolute right-4 top-4 rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:pointer-events-none">
        <X className="h-6 w-6" />
        <span className="sr-only">Close</span>
      </SheetClosePrimitive>
    </SheetContentPrimitive>
  </SheetPortal>
))
SheetContent.displayName = 'SheetContent'

const SheetHeader = ({ className, ...props }: HTMLAttributes<HTMLDivElement>) => (
  <div className={cn('flex flex-col space-y-1.5', className)} {...props} />
)
SheetHeader.displayName = 'SheetHeader'

const SheetFooter = ({ className, ...props }: HTMLAttributes<HTMLDivElement>) => (
  <div className={cn('flex flex-col-reverse sm:flex-row sm:justify-end sm:space-x-2', className)} {...props} />
)
SheetFooter.displayName = 'SheetFooter'

const SheetTitle = forwardRef<
  ComponentRef<typeof SheetTitlePrimitive>,
  ComponentPropsWithoutRef<typeof SheetTitlePrimitive>
>(({ className, ...props }, ref) => (
  <SheetTitlePrimitive ref={ref} className={cn('text-lg font-semibold text-foreground', className)} {...props} />
))
SheetTitle.displayName = 'SheetTitle'

const SheetDescription = forwardRef<
  ComponentRef<typeof SheetDescriptionPrimitive>,
  ComponentPropsWithoutRef<typeof SheetDescriptionPrimitive>
>(({ className, ...props }, ref) => (
  <SheetDescriptionPrimitive
    ref={ref}
    className={cn('text-sm text-muted-foreground', className)}
    {...props}
  />
))
SheetDescription.displayName = 'SheetDescription'

export { Sheet, SheetTrigger, SheetClose, SheetContent, SheetHeader, SheetFooter, SheetTitle, SheetDescription }
