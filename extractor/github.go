package extractor

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// extractGitHub handles GitHub issues, PRs, and repo READMEs.
func extractGitHub(html string, u *url.URL) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", err
	}

	path := strings.Trim(u.Path, "/")
	parts := strings.Split(path, "/")

	// Detect what kind of GitHub page this is
	switch {
	case len(parts) >= 4 && (parts[2] == "issues" || parts[2] == "pull"):
		return extractGitHubIssue(doc, parts)
	case len(parts) >= 2:
		return extractGitHubRepo(doc, parts)
	default:
		return "", fmt.Errorf("unrecognized GitHub URL pattern")
	}
}

func extractGitHubIssue(doc *goquery.Document, parts []string) (string, error) {
	var sb strings.Builder

	// Title
	title := strings.TrimSpace(doc.Find(".gh-header-title .js-issue-title").First().Text())
	if title == "" {
		title = strings.TrimSpace(doc.Find("[data-hpc] bdi").First().Text())
	}
	if title == "" {
		title = strings.TrimSpace(doc.Find("h1").First().Text())
	}

	kind := "Issue"
	if parts[2] == "pull" {
		kind = "Pull Request"
	}

	number := ""
	if len(parts) >= 4 {
		number = "#" + parts[3]
	}

	sb.WriteString(fmt.Sprintf("## %s %s: %s\n", kind, number, title))
	sb.WriteString(fmt.Sprintf("Repository: %s/%s\n\n", parts[0], parts[1]))

	// State (open/closed/merged)
	state := strings.TrimSpace(doc.Find(".State").First().Text())
	if state == "" {
		state = strings.TrimSpace(doc.Find("[title='Status: Open'], [title='Status: Closed'], [title='Status: Merged']").First().AttrOr("title", ""))
	}
	if state != "" {
		sb.WriteString(fmt.Sprintf("Status: %s\n\n", state))
	}

	// Labels
	var labels []string
	doc.Find(".IssueLabel, .label").Each(func(_ int, s *goquery.Selection) {
		label := strings.TrimSpace(s.Text())
		if label != "" {
			labels = append(labels, label)
		}
	})
	if len(labels) > 0 {
		sb.WriteString("Labels: " + strings.Join(labels, ", ") + "\n\n")
	}

	// Comments (issue body + discussion)
	sb.WriteString("---\n\n")
	doc.Find(".timeline-comment, .comment").Each(func(i int, s *goquery.Selection) {
		author := strings.TrimSpace(s.Find(".author, .timeline-comment-header strong").First().Text())
		body := strings.TrimSpace(s.Find(".comment-body, .markdown-body").First().Text())

		if body == "" {
			return
		}

		if i == 0 {
			sb.WriteString(fmt.Sprintf("**%s** (OP):\n", author))
		} else {
			sb.WriteString(fmt.Sprintf("\n**%s**:\n", author))
		}
		sb.WriteString(cleanText(body))
		sb.WriteString("\n\n---\n\n")
	})

	result := sb.String()
	if len(result) < 50 {
		return "", fmt.Errorf("extracted too little content")
	}
	return result, nil
}

func extractGitHubRepo(doc *goquery.Document, parts []string) (string, error) {
	var sb strings.Builder

	repo := strings.Join(parts[:2], "/")
	sb.WriteString(fmt.Sprintf("## %s\n\n", repo))

	// Description
	desc := strings.TrimSpace(doc.Find("[itemprop='about'], .f4.my-3").First().Text())
	if desc != "" {
		sb.WriteString(desc + "\n\n")
	}

	// README content
	readme := strings.TrimSpace(doc.Find("#readme .markdown-body, article.markdown-body").First().Text())
	if readme != "" {
		sb.WriteString("---\n\n")
		sb.WriteString(cleanText(readme))
		sb.WriteString("\n")
	}

	result := sb.String()
	if len(result) < 30 {
		return "", fmt.Errorf("extracted too little content")
	}
	return result, nil
}
