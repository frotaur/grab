package extractor

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// extractStackOverflow handles Stack Overflow question pages.
func extractStackOverflow(html string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", err
	}

	var sb strings.Builder

	// Question title
	title := strings.TrimSpace(doc.Find("#question-header h1").First().Text())
	if title == "" {
		title = strings.TrimSpace(doc.Find("h1").First().Text())
	}
	if title != "" {
		sb.WriteString("## " + title + "\n\n")
	}

	// Tags
	var tags []string
	doc.Find(".post-tag, .s-tag").Each(func(_ int, s *goquery.Selection) {
		tag := strings.TrimSpace(s.Text())
		if tag != "" {
			tags = append(tags, tag)
		}
	})
	if len(tags) > 0 {
		sb.WriteString("Tags: " + strings.Join(tags, ", ") + "\n\n")
	}

	// Vote count
	votes := strings.TrimSpace(doc.Find(".question .js-vote-count").First().Text())
	if votes != "" {
		sb.WriteString(fmt.Sprintf("Votes: %s\n\n", votes))
	}

	// Question body
	sb.WriteString("---\n\n")
	sb.WriteString("### Question\n\n")
	questionBody := strings.TrimSpace(doc.Find(".question .s-prose, .question .post-text").First().Text())
	if questionBody != "" {
		sb.WriteString(cleanText(questionBody))
		sb.WriteString("\n\n")
	}

	// Answers
	sb.WriteString("---\n\n")
	answers := doc.Find(".answer")
	if answers.Length() == 0 {
		sb.WriteString("No answers yet.\n")
		return sb.String(), nil
	}

	answers.Each(func(i int, s *goquery.Selection) {
		// Only include top 3 answers to keep output manageable
		if i >= 3 {
			return
		}

		isAccepted := s.HasClass("accepted-answer")
		votes := strings.TrimSpace(s.Find(".js-vote-count").First().Text())

		header := fmt.Sprintf("### Answer %d", i+1)
		if isAccepted {
			header += " [ACCEPTED]"
		}
		if votes != "" {
			header += fmt.Sprintf(" (%s votes)", votes)
		}
		sb.WriteString(header + "\n\n")

		body := strings.TrimSpace(s.Find(".s-prose, .post-text").First().Text())
		if body != "" {
			sb.WriteString(cleanText(body))
			sb.WriteString("\n\n---\n\n")
		}
	})

	remaining := answers.Length() - 3
	if remaining > 0 {
		sb.WriteString(fmt.Sprintf("[%d more answers not shown]\n", remaining))
	}

	result := sb.String()
	if len(result) < 50 {
		return "", fmt.Errorf("extracted too little content")
	}
	return result, nil
}
