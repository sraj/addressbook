import { useState } from 'react'
import FormSheet from '@/components/ui/FormSheet'
import { Button } from '@/components/ui/button'
import { toast } from '@/hooks/use-toast'
import { showErrorToast } from '@/lib/error'
import { api } from '@/lib/api'
import { parseURLList, type ParsedBookmark } from '@/lib/bookmark-import'

interface BookmarkImportDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onImport: (bookmarks: ParsedBookmark[]) => Promise<void>
}

type Tab = 'paste' | 'upload'

export default function BookmarkImportDialog({ open, onOpenChange, onImport }: BookmarkImportDialogProps) {
  const [tab, setTab] = useState<Tab>('paste')
  const [text, setText] = useState('')
  const [parsed, setParsed] = useState<ParsedBookmark[]>([])
  const [importing, setImporting] = useState(false)
  const [result, setResult] = useState<{ imported: number; skipped: number } | null>(null)

  const handleParse = () => {
    if (tab === 'paste') {
      setParsed(parseURLList(text))
    }
  }

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    const reader = new FileReader()
    reader.onload = async () => {
      const html = reader.result as string
      setImporting(true)
      try {
        const res = await api.importBookmarkHTML(html)
        setResult(res)
        toast({ title: `Imported ${res.imported} bookmarks`, variant: 'success' })
        onOpenChange(false)
      } catch (err) {
        showErrorToast(err, 'Import failed')
      } finally {
        setImporting(false)
      }
    }
    reader.readAsText(file)
  }

  const handleImport = async () => {
    if (parsed.length === 0) return
    setImporting(true)
    try {
      await onImport(parsed)
      toast({ title: 'Bookmarks imported', variant: 'success' })
      onOpenChange(false)
    } catch (err) {
      showErrorToast(err, 'Import failed')
    } finally {
      setImporting(false)
    }
  }

  return (
    <FormSheet
      open={open}
      onOpenChange={onOpenChange}
      title="Import Bookmarks"
      footer={
        tab === 'paste' ? (
          <>
            <Button variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
            <Button onClick={handleImport} disabled={parsed.length === 0 || importing}>
              {importing ? 'Importing…' : `Import ${parsed.length} bookmark${parsed.length !== 1 ? 's' : ''}`}
            </Button>
          </>
        ) : undefined
      }
    >
      <div className="space-y-4">
        <div className="flex items-center gap-2 border-b">
            <button
              type="button"
              onClick={() => { setTab('paste'); setParsed([]); setResult(null) }}
              className={`px-4 py-2 text-sm font-medium border-b-2 -mb-[1px] transition-colors ${
                tab === 'paste' ? 'border-primary text-foreground' : 'border-transparent text-muted-foreground hover:text-foreground'
              }`}
            >
              Paste URLs
            </button>
            <button
              type="button"
              onClick={() => { setTab('upload'); setParsed([]); setResult(null) }}
              className={`px-4 py-2 text-sm font-medium border-b-2 -mb-[1px] transition-colors ${
                tab === 'upload' ? 'border-primary text-foreground' : 'border-transparent text-muted-foreground hover:text-foreground'
              }`}
            >
              Upload HTML
            </button>
          </div>

        {tab === 'paste' && (
          <div className="space-y-3">
            <p className="text-sm text-muted-foreground">
              Paste one URL per line. Optionally add a title after a pipe: <code>url | Title</code>
            </p>
            <textarea
              className="flex min-h-[200px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring font-mono"
              placeholder={`https://example.com\nhttps://github.com | GitHub\nhttps://opencode.ai`}
              value={text}
              onChange={(e) => setText(e.target.value)}
            />
            <Button onClick={handleParse} disabled={!text.trim()}>
              Preview
            </Button>
          </div>
        )}

        {tab === 'upload' && (
          <div className="space-y-3">
            <p className="text-sm text-muted-foreground">
              Upload your browser's bookmark export HTML file. The file will be parsed on the server.
            </p>
            <input
              type="file"
              accept=".html,.htm"
              onChange={handleFileUpload}
              disabled={importing}
              className="block w-full text-sm text-muted-foreground file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-sm file:font-medium file:bg-primary/10 file:text-primary hover:file:bg-primary/20"
            />
            {importing && <p className="text-sm text-muted-foreground">Processing file…</p>}
            {result && (
              <p className="text-sm text-muted-foreground">
                Imported {result.imported} bookmark{result.imported !== 1 ? 's' : ''}
                {result.skipped > 0 && ` (${result.skipped} skipped as duplicates)`}
              </p>
            )}
          </div>
        )}

        {tab === 'paste' && parsed.length > 0 && (
          <div className="space-y-2">
            <p className="text-sm font-medium">
              {parsed.length} bookmark{parsed.length !== 1 ? 's' : ''} found
            </p>
            <div className="max-h-[300px] overflow-y-auto space-y-1 rounded-md border p-2">
              {parsed.slice(0, 50).map((b, i) => (
                <div key={i} className="flex items-center gap-2 text-sm truncate">
                  <span className="text-muted-foreground shrink-0 w-6 text-right">{i + 1}.</span>
                  <span className="font-medium truncate">{b.title}</span>
                  {b.category && <span className="text-xs text-muted-foreground shrink-0">({b.category})</span>}
                </div>
              ))}
              {parsed.length > 50 && (
                <p className="text-xs text-muted-foreground text-center pt-1">
                  ...and {parsed.length - 50} more
                </p>
              )}
            </div>
          </div>
        )}
      </div>
    </FormSheet>
  )
}
