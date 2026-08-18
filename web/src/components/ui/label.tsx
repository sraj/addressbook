import { Label as NLabel, type LabelProps as NLabelProps } from '@mobentum/nebula-ui'

export type LabelProps = NLabelProps

const Label = (props: LabelProps) => {
  return <NLabel {...props} />
}
Label.displayName = 'Label'

export { Label }