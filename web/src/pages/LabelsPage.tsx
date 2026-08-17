import PageLayout from '@/components/layout/PageLayout'
import PageHeader from '@/components/layout/PageHeader'
import LabelPrintOrderCard from '@/components/labels/LabelPrintOrderCard'

export default function LabelsPage() {
  return (
    <PageLayout>
      <PageHeader
        title="Address Labels"
        description="Download a print sheet or order a set"
      />

      <div className="max-w-4xl">
        <LabelPrintOrderCard />
      </div>
    </PageLayout>
  )
}