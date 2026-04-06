package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/frotaur/grab/extractor"
	"github.com/frotaur/grab/fetcher"
	"github.com/spf13/cobra"
)

var (
	maxTokens  int
	listLinks  bool
	rawMode    bool
	noCache    bool
	userAgent  string

	rootCmd = &cobra.Command{
		Use:   "grab [URL]",
		Short: "Headless web reader for terminals and AI agents.",
		Long: `grab fetches a URL, extracts the main content, and outputs
clean plain text. Built for terminals and AI agents that need
to read web pages without drowning in HTML noise.

Examples:
  grab https://docs.stripe.com/api/charges
  grab https://github.com/user/repo/issues/42
  grab https://stackoverflow.com/questions/12345
  grab --tokens 2000 https://docs.python.org/3/library/json.html
  grab --list https://docs.foo.com/api`,
		Args: cobra.ExactArgs(1),
		RunE: run,
	}
)

func init() {
	rootCmd.Flags().IntVarP(&maxTokens, "tokens", "t", 0, "truncate output to approximately this many tokens (0 = no limit)")
	rootCmd.Flags().BoolVarP(&listLinks, "list", "l", false, "list links found on the page instead of content")
	rootCmd.Flags().BoolVarP(&rawMode, "raw", "r", false, "output raw extracted text without cleaning")
	rootCmd.Flags().BoolVar(&noCache, "no-cache", false, "bypass the local cache")
	rootCmd.Flags().StringVar(&userAgent, "user-agent", "grab/1.0 (headless web reader)", "HTTP User-Agent header")
}

func Execute() error {
	return rootCmd.Execute()
}

func run(cmd *cobra.Command, args []string) error {
	url := args[0]

	// Ensure URL has a scheme
	if !strings.Contains(url, "://") {
		url = "https://" + url
	}

	// Fetch the page
	html, finalURL, err := fetcher.Fetch(url, userAgent, !noCache)
	if err != nil {
		return fmt.Errorf("failed to fetch %s: %w", url, err)
	}

	// List links mode
	if listLinks {
		links, err := extractor.ExtractLinks(strings.NewReader(html), finalURL)
		if err != nil {
			return fmt.Errorf("failed to extract links: %w", err)
		}
		for _, link := range links {
			fmt.Fprintf(os.Stdout, "%s\n  %s\n", link.Text, link.URL)
		}
		return nil
	}

	// Extract content
	content, err := extractor.Extract(strings.NewReader(html), finalURL)
	if err != nil {
		return fmt.Errorf("failed to extract content: %w", err)
	}

	// Token limiting
	if maxTokens > 0 {
		content = truncateToTokens(content, maxTokens)
	}

	_, err = io.WriteString(os.Stdout, content)
	return err
}

// truncateToTokens does a rough truncation based on ~4 chars per token.
// It tries to break at paragraph boundaries.
func truncateToTokens(s string, tokens int) string {
	maxChars := tokens * 4
	if len(s) <= maxChars {
		return s
	}

	// Try to break at a paragraph boundary
	truncated := s[:maxChars]
	if idx := strings.LastIndex(truncated, "\n\n"); idx > maxChars/2 {
		truncated = truncated[:idx]
	} else if idx := strings.LastIndex(truncated, "\n"); idx > maxChars*3/4 {
		truncated = truncated[:idx]
	}

	return truncated + "\n\n[... truncated to ~" + fmt.Sprintf("%d", tokens) + " tokens]\n"
}
