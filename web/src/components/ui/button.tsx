import { Button as NButton, buttonVariants as nbuttonVariants, type ButtonProps as NButtonProps } from '@mobentum/nebula-ui'

export { nbuttonVariants as buttonVariants }

export type ButtonProps = NButtonProps

const Button = (props: ButtonProps) => {
  return <NButton {...props} />
}
Button.displayName = 'Button'

export { Button }