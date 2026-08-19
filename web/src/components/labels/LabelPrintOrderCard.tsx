import { useEffect, useMemo, useState } from 'react'
import { useLabelsStore } from '@/store/labels'
import { useCollectionsStore } from '@/store/collections'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import EmptyState from '@/components/ui/EmptyState'
import { toast } from '@/hooks/use-toast'
import { showErrorToast } from '@/lib/error'
import { api } from '@/lib/api'
import { Printer, CreditCard, Download, Tag } from 'lucide-react'

interface LabelPrintOrderCardProps {
  collectionId?: number
}

function formatAmount(cents: number, currency: string) {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: currency.toUpperCase() }).format(cents / 100)
}

function formatDate(s?: string) {
  if (!s) return ''
  return new Date(s).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' })
}

const statusColors: Record<string, string> = {
  pending: 'bg-muted text-muted-foreground',
  paid: 'bg-success/10 text-success',
  canceled: 'bg-destructive/10 text-destructive',
}

export default function LabelPrintOrderCard({ collectionId }: LabelPrintOrderCardProps) {
  const { orders, formats, loading, fetchOrders, fetchFormats, createOrder } = useLabelsStore()
  const { collections, fetchCollections } = useCollectionsStore()

  const [scope, setScope] = useState(collectionId ? String(collectionId) : 'all')
  const [format, setFormat] = useState('5160')
  const [email, setEmail] = useState('')
  const [creating, setCreating] = useState(false)
  const [downloading, setDownloading] = useState(false)

  useEffect(() => {
    fetchFormats()
    fetchCollections()
    fetchOrders()
  }, [])

  useEffect(() => {
    if (collectionId) setScope(String(collectionId))
  }, [collectionId])

  const effectiveCollectionId = scope === 'all' ? undefined : Number(scope)

  const scopedOrders = useMemo(() => {
    if (!collectionId) return orders
    return orders.filter((o) => o.collection_id === collectionId)
  }, [orders, collectionId])

  const handleDownload = async () => {
    setDownloading(true)
    try {
      await api.downloadLabelSheet(effectiveCollectionId, format)
    } catch (err) {
      showErrorToast(err, 'Failed to download label sheet')
    } finally {
      setDownloading(false)
    }
  }

  const handleOrder = async () => {
    setCreating(true)
    try {
      const res = await createOrder(effectiveCollectionId, email.trim() || undefined, format)
      if (res.url) {
        window.location.href = res.url
      } else {
        toast({ title: 'Order placed', description: 'Print your labels from the sheet below', variant: 'success' })
      }
    } catch (err) {
      showErrorToast(err, 'Failed to create label order')
    } finally {
      setCreating(false)
    }
  }

  return (
    <div className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="label-format">Label format</Label>
        <Select value={format} onValueChange={setFormat}>
          <SelectTrigger id="label-format" className="w-full">
            <SelectValue placeholder="Select a format" />
          </SelectTrigger>
          <SelectContent>
            {formats.map((f) => (
              <SelectItem key={f.code} value={f.code}>{f.name}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10">
                <Printer className="h-4 w-4 text-primary" />
              </div>
              <div>
                <CardTitle className="text-base">Print Labels</CardTitle>
                <CardDescription>Download a sheet and print it yourself</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent className="space-y-3">
            {!collectionId && (
              <div className="space-y-2">
                <Label htmlFor="label-scope">Contacts</Label>
                <Select value={scope} onValueChange={setScope}>
                  <SelectTrigger id="label-scope" className="w-full">
                    <SelectValue placeholder="Select a collection" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All contacts</SelectItem>
                    {collections.map((c) => (
                      <SelectItem key={c.id} value={String(c.id)}>{c.name} ({c.contact_count ?? 0})</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            )}
            <div className="flex flex-wrap gap-2">
              <Button variant="outline" className="flex-1 min-w-[140px]" onClick={handleDownload} disabled={downloading}>
                <Download className="h-4 w-4" />
                {downloading ? 'Downloading…' : 'Download Sheet'}
              </Button>
              <a href={`/api/v1/labels/sheet${effectiveCollectionId ? `?collection_id=${effectiveCollectionId}` : ''}${format ? `${effectiveCollectionId ? '&' : '?'}format=${format}` : ''}`} target="_blank" rel="noopener noreferrer" className="flex-1 min-w-[140px]">
                <Button variant="outline" className="w-full">
                  <Printer className="h-4 w-4" />
                  Open Print Sheet
                </Button>
              </a>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10">
                <Tag className="h-4 w-4 text-primary" />
              </div>
              <div>
                <CardTitle className="text-base">Order Labels</CardTitle>
                <CardDescription>We print and ship them to you</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent className="space-y-3">
            {!collectionId && (
              <div className="space-y-2">
                <Label htmlFor="order-scope">Contacts</Label>
                <Select value={scope} onValueChange={setScope}>
                  <SelectTrigger id="order-scope" className="w-full">
                    <SelectValue placeholder="Select a collection" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All contacts</SelectItem>
                    {collections.map((c) => (
                      <SelectItem key={c.id} value={String(c.id)}>{c.name} ({c.contact_count ?? 0})</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            )}
            <div className="space-y-2">
              <Label htmlFor="order-email">Email for receipt</Label>
              <Input id="order-email" type="email" value={email} onChange={(e) => setEmail(e.target.value)} placeholder="you@example.com" />
            </div>
            <Button className="w-full" onClick={handleOrder} disabled={creating}>
              <CreditCard className="h-4 w-4" />
              {creating ? 'Creating checkout…' : 'Order labels'}
            </Button>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <Tag className="h-4 w-4" />
            Order History
          </CardTitle>
          <CardDescription>Labels you've ordered</CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <p className="text-sm text-muted-foreground">Loading…</p>
          ) : scopedOrders.length === 0 ? (
            <EmptyState
              icon={<Tag className="h-10 w-10" />}
              heading="No orders yet"
              description="Order a set of printed labels to see your history"
            />
          ) : (
            <div className="rounded-lg border bg-background">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Date</TableHead>
                    <TableHead>Format</TableHead>
                    <TableHead>Contacts</TableHead>
                    <TableHead>Sheets</TableHead>
                    <TableHead>Amount</TableHead>
                    <TableHead className="text-right">Status</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {scopedOrders.map((order) => (
                    <TableRow key={order.id}>
                      <TableCell className="text-muted-foreground">{formatDate(order.created_at)}</TableCell>
                      <TableCell>{order.label_type || '5160'}</TableCell>
                      <TableCell>{order.contact_count}</TableCell>
                      <TableCell>{order.sheet_count}</TableCell>
                      <TableCell>{formatAmount(order.amount_cents, order.currency)}</TableCell>
                      <TableCell className="text-right">
                        <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium capitalize ${statusColors[order.status] || 'bg-muted text-muted-foreground'}`}>
                          {order.status}
                        </span>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}