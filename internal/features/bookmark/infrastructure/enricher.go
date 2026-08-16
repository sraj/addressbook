package infrastructure

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

func FaviconURLFromURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("https://www.google.com/s2/favicons?domain=%s&sz=64", parsed.Hostname())
}

type EnrichedData struct {
	Title       string
	Description string
	FaviconURL  string
}

func EnrichURL(rawURL string) (*EnrichedData, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	data := &EnrichedData{}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(rawURL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	// Limit reading to 1MB
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}

	data.Title = extractTitle(body)
	data.Description = extractMetaDescription(body)

	// Favicon: prefer og:image or apple-touch-icon, fallback to /favicon.ico
	data.FaviconURL = extractFavicon(body, parsed)
	if data.FaviconURL == "" {
		data.FaviconURL = parsed.Scheme + "://" + parsed.Host + "/favicon.ico"
	}

	return data, nil
}

func extractTitle(body []byte) string {
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return ""
	}

	var findTitle func(*html.Node) string
	findTitle = func(n *html.Node) string {
		if n.Type == html.ElementNode && n.Data == "title" && n.FirstChild != nil {
			return strings.TrimSpace(n.FirstChild.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if t := findTitle(c); t != "" {
				return t
			}
		}
		return ""
	}

	return findTitle(doc)
}

func extractMetaDescription(body []byte) string {
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return ""
	}

	var findDesc func(*html.Node) string
	findDesc = func(n *html.Node) string {
		if n.Type == html.ElementNode && n.Data == "meta" {
			var name, content string
			for _, attr := range n.Attr {
				switch attr.Key {
				case "name":
					name = strings.ToLower(attr.Val)
				case "property":
					if strings.ToLower(attr.Val) == "og:description" {
						name = "description"
					}
				case "content":
					content = strings.TrimSpace(attr.Val)
				}
			}
			if name == "description" && content != "" {
				return content
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if d := findDesc(c); d != "" {
				return d
			}
		}
		return ""
	}

	return findDesc(doc)
}

func extractFavicon(body []byte, base *url.URL) string {
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return ""
	}

	var findIcon func(*html.Node) string
	findIcon = func(n *html.Node) string {
		if n.Type == html.ElementNode && n.Data == "link" {
			var rel, href string
			for _, attr := range n.Attr {
				switch attr.Key {
				case "rel":
					rel = strings.ToLower(attr.Val)
				case "href":
					href = attr.Val
				}
			}
			if href != "" && (rel == "icon" || rel == "shortcut icon" || rel == "apple-touch-icon") {
				iconURL, err := url.Parse(href)
				if err == nil {
					if !iconURL.IsAbs() {
						iconURL = base.ResolveReference(iconURL)
					}
					return iconURL.String()
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if i := findIcon(c); i != "" {
				return i
			}
		}
		return ""
	}

	return findIcon(doc)
}

type ParsedBookmark struct {
	URL        string
	Title      string
	Category   string
	FaviconURL string
}

func ParseBookmarkHTML(htmlContent string) []ParsedBookmark {
	var results []ParsedBookmark
	currentCategory := ""

	for _, line := range strings.Split(htmlContent, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for folder: <DT><H3 ...>Folder Name</H3>
		if idx := strings.Index(strings.ToUpper(line), "<H3"); idx >= 0 {
			endTag := strings.Index(line, "</H3>")
			if endTag > idx {
				content := line[idx+3 : endTag]
				if gt := strings.Index(content, ">"); gt >= 0 {
					currentCategory = strings.TrimSpace(content[gt+1:])
				}
				continue
			}
		}

		// Check for bookmark: <DT><A HREF="url" ...>Title</A>
		aIdx := strings.Index(strings.ToUpper(line), "<A ")
		if aIdx < 0 {
			continue
		}

		hrefPart := line[aIdx+3:]
		hrefStart := strings.Index(hrefPart, `HREF="`)
		if hrefStart < 0 {
			continue
		}
		hrefStart += 6 // len of `HREF="`
		hrefEnd := strings.Index(hrefPart[hrefStart:], `"`)
		if hrefEnd < 0 {
			continue
		}
		url := hrefPart[hrefStart : hrefStart+hrefEnd]

		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			continue
		}

		// Extract title between > and </A>
		titleStart := strings.LastIndex(hrefPart[:hrefStart+hrefEnd+1], ">")
		var title string
		if titleStart >= 0 {
			titleContent := hrefPart[titleStart+1:]
			titleEnd := strings.Index(titleContent, "</A>")
			if titleEnd >= 0 {
				title = strings.TrimSpace(titleContent[:titleEnd])
			}
		}
		if title == "" {
			title = url
		}

		results = append(results, ParsedBookmark{
			URL:        url,
			Title:      title,
			Category:   currentCategory,
			FaviconURL: url, // will be enriched later
		})
	}

	return results
}
