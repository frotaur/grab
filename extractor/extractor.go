package extractor

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	readability "github.com/go-shiori/go-readability"
)

// Extract extracts the main content from HTML, returning clean plain text.
// It detects site-specific extractors for known sites and falls back to
// Mozilla's Readability algorithm for everything else.
func Extract(r io.Reader, pageURL string) (string, error) {
	u, err := url.Parse(pageURL)
	if err != nil {
		return "", fmt.Errorf("parsing URL: %w", err)
	}

	host := strings.ToLower(u.Hostname())

	// Read the HTML once
	htmlBytes, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("reading HTML: %w", err)
	}
	html := string(htmlBytes)

	// Try site-specific extractors
	switch {
	case host == "github.com":
		if content, err := extractGitHub(html, u); err == nil && content != "" {
			return content, nil
		}
	case host == "stackoverflow.com" || host == "www.stackoverflow.com" ||
		strings.HasSuffix(host, ".stackexchange.com"):
		if content, err := extractStackOverflow(html); err == nil && content != "" {
			return content, nil
		}
	}

	// Generic readability fallback
	return extractReadability(strings.NewReader(html), pageURL)
}

func extractReadability(r io.Reader, pageURL string) (string, error) {
	u, err := url.Parse(pageURL)
	if err != nil {
		return "", err
	}

	article, err := readability.FromReader(r, u)
	if err != nil {
		return "", fmt.Errorf("readability extraction failed: %w", err)
	}

	var sb strings.Builder

	if article.Title != "" {
		sb.WriteString("## ")
		sb.WriteString(article.Title)
		sb.WriteString("\n\n")
	}

	// Use the text content — readability gives us clean text
	content := article.TextContent
	if content == "" {
		content = article.Content // fallback to HTML content
	}

	// Clean up excessive whitespace
	content = cleanText(content)
	sb.WriteString(content)
	sb.WriteString("\n")

	return sb.String(), nil
}

// cleanText normalizes whitespace in extracted text.
func cleanText(s string) string {
	// Collapse runs of 3+ newlines to 2
	for strings.Contains(s, "\n\n\n") {
		s = strings.ReplaceAll(s, "\n\n\n", "\n\n")
	}

	// Trim leading/trailing whitespace from each line
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	s = strings.Join(lines, "\n")

	return strings.TrimSpace(s)
}
