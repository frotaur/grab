package fetcher

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const cacheTTL = 1 * time.Hour

func cacheDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".grab", "cache")
}

func cacheKey(url string) string {
	h := sha256.Sum256([]byte(url))
	return hex.EncodeToString(h[:])
}

// Fetch retrieves the HTML content of a URL.
// If useCache is true, it will check/store results in ~/.grab/cache/.
func Fetch(url, userAgent string, useCache bool) (html string, finalURL string, err error) {
	if useCache {
		if cached, ok := readCache(url); ok {
			return cached, url, nil
		}
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("fetching URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("reading response: %w", err)
	}

	html = string(body)
	finalURL = resp.Request.URL.String()

	if useCache {
		writeCache(url, html)
	}

	return html, finalURL, nil
}

func readCache(url string) (string, bool) {
	dir := cacheDir()
	if dir == "" {
		return "", false
	}
	key := cacheKey(url)
	path := filepath.Join(dir, key)

	info, err := os.Stat(path)
	if err != nil {
		return "", false
	}

	// Check TTL
	if time.Since(info.ModTime()) > cacheTTL {
		os.Remove(path)
		return "", false
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}

	return string(data), true
}

func writeCache(url, html string) {
	dir := cacheDir()
	if dir == "" {
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	key := cacheKey(url)
	path := filepath.Join(dir, key)
	_ = os.WriteFile(path, []byte(html), 0o644)
}
