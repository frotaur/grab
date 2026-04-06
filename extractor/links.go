package extractor

import (
	"io"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Link represents an extracted link with its text and URL.
type Link struct {
	Text string
	URL  string
}

// ExtractLinks extracts all meaningful links from an HTML page.
// It filters out navigation noise, empty links, and anchors.
func ExtractLinks(r io.Reader, pageURL string) ([]Link, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	base, _ := url.Parse(pageURL)
	seen := make(map[string]bool)
	var links []Link

	// Remove nav, header, footer elements to reduce noise
	doc.Find("nav, header, footer, .nav, .navbar, .footer, .header, .sidebar").Remove()

	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists || href == "" || href == "#" {
			return
		}

		// Skip javascript: and mailto: links
		if strings.HasPrefix(href, "javascript:") || strings.HasPrefix(href, "mailto:") {
			return
		}

		// Resolve relative URLs
		resolved, err := base.Parse(href)
		if err != nil {
			return
		}
		fullURL := resolved.String()

		// Skip duplicates
		if seen[fullURL] {
			return
		}
		seen[fullURL] = true

		// Get link text, clean it up
		text := strings.TrimSpace(s.Text())
		if text == "" {
			// Try alt text from images
			text = strings.TrimSpace(s.Find("img").AttrOr("alt", ""))
		}
		if text == "" {
			text = "[no text]"
		}

		// Collapse whitespace
		fields := strings.Fields(text)
		text = strings.Join(fields, " ")

		// Skip very short meaningless link text
		if len(text) < 2 && text != "[no text]" {
			return
		}

		links = append(links, Link{Text: text, URL: fullURL})
	})

	return links, nil
}
