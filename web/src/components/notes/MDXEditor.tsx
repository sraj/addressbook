import { useState, useRef, useCallback, useEffect } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { HelpCircle } from 'lucide-react'

interface MDXEditorProps {
  value: string
  onChange: (value: string) => void
  editable?: boolean
  title?: string
}

// MDX syntax cheatsheet entries
const syntaxHelp = [
  ['# Heading', '## Heading', '### Heading'],
  ['**Bold**', '*Italic*', '~~Strikethrough~~'],
  ['[Link](url)', '![Image](url)'],
  ['`Inline code`', '```\nCode block\n```'],
  ['- List item', '1. Numbered item'],
  ['> Blockquote', '--- (horizontal rule)'],
  ['| Col1 | Col2 |', '|------|------|', '| A    | B    |'],
]

export default function MDXEditor({
  value,
  onChange,
  editable = true,
  title,
}: MDXEditorProps) {
  // Tab and help popover state
  const [tab, setTab] = useState<'edit' | 'preview'>('edit')
  const [showHelp, setShowHelp] = useState(false)
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const gutterRef = useRef<HTMLDivElement>(null)
  const helpRef = useRef<HTMLDivElement>(null)

  // Line count for gutter numbering
  const lineCount = value.split('\n').length
  const gutterDigits = String(lineCount).length

  // Auto-resize textarea height to fit content
  const autoResize = useCallback(() => {
    const ta = textareaRef.current
    if (!ta) return
    ta.style.height = '0px'
    ta.style.height = ta.scrollHeight + 'px'
  }, [])

  useEffect(() => {
    autoResize()
  }, [value, autoResize])

  // Close help popover on outside click
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (helpRef.current && !helpRef.current.contains(e.target as Node)) {
        setShowHelp(false)
      }
    }
    if (showHelp) {
      document.addEventListener('mousedown', handleClickOutside)
      return () => document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [showHelp])

  const handleChange = useCallback(
    (e: React.ChangeEvent<HTMLTextAreaElement>) => {
      onChange(e.target.value)
    },
    [onChange],
  )

  // Sync gutter scroll with textarea scroll
  const handleScroll = useCallback(() => {
    if (textareaRef.current && gutterRef.current) {
      gutterRef.current.scrollTop = textareaRef.current.scrollTop
    }
  }, [])

  // Tab key inserts 2 spaces instead of moving focus
  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
      if (e.key === 'Tab') {
        e.preventDefault()
        const ta = e.currentTarget
        const start = ta.selectionStart
        const end = ta.selectionEnd
        const before = value.slice(0, start)
        const after = value.slice(end)
        onChange(before + '  ' + after)
        requestAnimationFrame(() => {
          ta.selectionStart = ta.selectionEnd = start + 2
        })
      }
    },
    [value, onChange],
  )

  // Dynamic class for active/inactive tab button
  const tabClass = (name: 'edit' | 'preview') =>
    `px-4 py-1.5 text-sm font-medium transition-colors border-b-2 -mb-[1px] ${
      tab === name
        ? 'border-primary bg-background text-foreground'
        : 'border-transparent text-muted-foreground hover:text-foreground'
    }`

  return (
    <div className="rounded-md border flex flex-col flex-1">
      {/* Header bar with tabs and help */}
      <div className="flex items-center border-b bg-muted/50 shrink-0 relative">
        {editable ? (
          <>
            <button type="button" onClick={() => setTab('edit')} className={tabClass('edit')}>
              Edit
            </button>
            <button type="button" onClick={() => setTab('preview')} className={tabClass('preview')}>
              Preview
            </button>
          </>
        ) : (
          <div className="px-4 py-1.5 text-md font-medium text-foreground">
            {title ?? 'Note'}
          </div>
        )}
        {editable && (
          <button
            type="button"
            onClick={() => setShowHelp(!showHelp)}
            className="ml-auto px-2 text-muted-foreground/50 hover:text-muted-foreground transition-colors"
            title="MDX syntax"
          >
            <HelpCircle className="h-4 w-4" />
          </button>
        )}
        {/* <span className="px-4 text-xs text-muted-foreground">.md</span> */}

        {/* MDX syntax help popover */}
        {showHelp && (
          <div
            ref={helpRef}
            className="absolute right-0 top-full z-50 mt-1 w-56 rounded-md border bg-popover p-3 shadow-md text-xs"
          >
            <p className="font-medium mb-2 text-foreground">MDX Syntax</p>
            <div className="space-y-1.5">
              {syntaxHelp.map((group, i) => (
                <div key={i} className="space-y-0.5">
                  {group.map((example) => (
                    <code key={example} className="block text-muted-foreground font-mono">
                      {example}
                    </code>
                  ))}
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* Edit mode: textarea with line-number gutter */}
      {editable && tab === 'edit' ? (
        <div className="flex flex-1 rounded-b-md">
          <div
            ref={gutterRef}
            className="select-none text-right px-2 py-2 text-sm leading-relaxed font-mono text-muted-foreground/40 bg-muted/30"
            style={{ minWidth: `${gutterDigits + 2}ch` }}
            aria-hidden="true"
          >
            {Array.from({ length: lineCount }, (_, i) => (
              <div key={i}>{i + 1}</div>
            ))}
          </div>
          <textarea
            ref={textareaRef}
            placeholder="Write your note in Markdown…"
            className="w-full border-0 bg-transparent px-3 py-2 text-sm leading-relaxed shadow-sm placeholder:text-muted-foreground/50 focus-visible:outline-none focus-visible:ring-0 font-mono resize-none overflow-hidden"
            value={value}
            onChange={handleChange}
            onScroll={handleScroll}
            onKeyDown={handleKeyDown}
            spellCheck={false}
          />
        </div>
      ) : (
        /* Preview mode: rendered markdown */
        <div className="prose dark:prose-invert max-w-none flex-1 rounded-b-md px-3 py-2 overflow-y-auto">
          {value ? (
            <ReactMarkdown
              remarkPlugins={[remarkGfm]}
              components={{
                code({ children, className, ...props }) {
                  if (className) {
                    return <code className={className} {...props}>{children}</code>
                  }
                  return (
                    <code className="rounded bg-muted px-1.5 py-0.5 text-sm font-mono" {...props}>
                      {children}
                    </code>
                  )
                },
              }}
            >{value}</ReactMarkdown>
          ) : (
            <p className="text-muted-foreground italic">Nothing to preview</p>
          )}
        </div>
      )}
    </div>
  )
}
