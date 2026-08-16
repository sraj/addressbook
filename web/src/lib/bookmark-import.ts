export interface ParsedBookmark {
  url: string
  title: string
  description?: string
  favicon_url?: string
  category?: string
}

export function extractDomain(url: string): string {
  try {
    return new URL(url).hostname.replace(/^www\./, '')
  } catch {
    return ''
  }
}

export function faviconURL(url: string): string {
  const domain = extractDomain(url)
  if (!domain) return ''
  return `https://www.google.com/s2/favicons?domain=${domain}&sz=64`
}

export function parseURLList(text: string): ParsedBookmark[] {
  const result: ParsedBookmark[] = []
  for (const line of text.split('\n')) {
    const trimmed = line.trim()
    if (!trimmed || trimmed.startsWith('#')) continue

    const match = trimmed.match(/^(https?:\/\/[^\s|]+)(?:\s*\|\s*(.+))?$/i)
    if (match) {
      const url = match[1].replace(/\/+$/, '')
      const title = match[2]?.trim() || url
      result.push({ url, title, favicon_url: faviconURL(url) })
      continue
    }

    const urlMatch = trimmed.match(/(https?:\/\/[^\s]+)/i)
    if (urlMatch) {
      const url = urlMatch[1].replace(/\/+$/, '')
      result.push({ url, title: url, favicon_url: faviconURL(url) })
    }
  }
  return result
}

export function parseBookmarkHTML(html: string): ParsedBookmark[] {
  const result: ParsedBookmark[] = []
  let currentCategory = ''
  // Normalize line endings
  const text = html.replace(/\r\n/g, '\n')

  // Split into lines for easier processing
  const lines = text.split('\n')

  for (const rawLine of lines) {
    const line = rawLine.trim()
    if (!line) continue

    // Check for folder: <DT><H3 ...>Folder Name</H3>
    const h3Match = line.match(/<H3[^>]*>(.*?)<\/H3>/i)
    if (h3Match) {
      currentCategory = h3Match[1].trim()
      continue
    }

    // Check for </DL> — pop the category stack on close
    // (Subfolders set category, but when they end, we restore parent)
    // We track nesting by counting <DL> and </DL>
    // Actually simpler: just use the LAST H3 seen as category

    // Check for bookmark: <DT><A HREF="url" ...>Title</A>
    const aMatch = line.match(/<A\s+HREF\s*=\s*"([^"]*)"[^>]*>([\s\S]*?)<\/A>/i)
    if (aMatch) {
      const url = aMatch[1].trim()
      const title = aMatch[2].replace(/<[^>]*>/g, '').trim()
      if (url && (url.startsWith('http://') || url.startsWith('https://'))) {
        result.push({
          url,
          title: title || url,
          favicon_url: faviconURL(url),
          category: currentCategory,
        })
      }
    }
  }

  return result
}
