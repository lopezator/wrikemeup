package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
)

// Client represents a GitHub client that can be used to interact with the GitHub API.
type Client struct {
	*http.Client
	repo     string
	botToken string
}

// NewClient creates a new Wrike client with the provided token.
func NewClient(botToken string, repo string) *Client {
	return &Client{
		Client:   &http.Client{},
		botToken: botToken,
		repo:     repo,
	}
}

// PostComment posts a comment on a GitHub issue using the GitHub bot.
func (c *Client) PostComment(issueNumber string) error {
	comment := "Hey! I just logged my hours on this task. Please check it out."
	return c.PostCommentWithBody(issueNumber, comment)
}

// PostCommentWithBody posts a custom comment on a GitHub issue.
func (c *Client) PostCommentWithBody(issueNumber string, comment string) error {
	commentURL := fmt.Sprintf("https://api.github.com/repos/%s/issues/%s/comments", c.repo, issueNumber)
	commentPayload := map[string]string{
		"body": comment,
	}
	commentBody, err := json.Marshal(commentPayload)
	if err != nil {
		return fmt.Errorf("wrikemeup: error when marshaling JSON contents: %w", err)
	}
	commentReq, err := http.NewRequest("POST", commentURL, bytes.NewBuffer(commentBody))
	if err != nil {
		return fmt.Errorf("error when creating the comment request: %w", err)
	}
	commentReq.Header.Set("Authorization", "Bearer "+c.botToken)
	commentReq.Header.Set("Accept", "application/vnd.github+json")
	commentReq.Header.Set("Content-Type", "application/json")
	commentResp, err := c.Do(commentReq)
	if err != nil {
		return fmt.Errorf("error when making the comment request: %w", err)
	}
	defer func() {
		if err := commentResp.Body.Close(); err != nil {
			log.Printf("wrikemeup: error closing response body: %v", err)
		}
	}()
	if commentResp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(commentResp.Body)
		return fmt.Errorf("wrikemeup: GitHub API error: %s", string(bodyBytes))
	}
	return nil
}

// PostHoursSummary posts a formatted summary of logged hours as a comment.
// Shows the complete current state of all logged hours.
func (c *Client) PostHoursSummary(issueNumber string, dailyHours map[string]float64, changes map[string]string) error {
	var summary strings.Builder
	summary.WriteString("## ✅ Hours Synced to Wrike\n\n")

	// If there are no hours, show that
	if len(dailyHours) == 0 {
		summary.WriteString("_No hours currently logged._\n")
		return c.PostCommentWithBody(issueNumber, summary.String())
	}

	summary.WriteString("### Current State\n")
	summary.WriteString("| Date | Hours | Status |\n")
	summary.WriteString("|------|-------|--------|\n")

	// Sort dates for consistent display
	dates := make([]string, 0, len(dailyHours))
	for date := range dailyHours {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	totalHours := 0.0
	for _, date := range dates {
		hours := dailyHours[date]
		totalHours += hours

		status := "✓"
		if change, ok := changes[date]; ok {
			status = change
		}

		summary.WriteString(fmt.Sprintf("| %s | %.2fh | %s |\n", date, hours, status))
	}

	// Check for deletions (dates in changes but not in dailyHours)
	deletedDates := make([]string, 0)
	for date, change := range changes {
		if _, exists := dailyHours[date]; !exists && strings.HasPrefix(change, "Deleted:") {
			deletedDates = append(deletedDates, date)
		}
	}

	if len(deletedDates) > 0 {
		sort.Strings(deletedDates)
		summary.WriteString("\n### Deleted Entries\n")
		for _, date := range deletedDates {
			summary.WriteString(fmt.Sprintf("| %s | - | %s |\n", date, changes[date]))
		}
	}

	summary.WriteString(fmt.Sprintf("\n**Total: %.2fh**\n", totalHours))

	return c.PostCommentWithBody(issueNumber, summary.String())
}

// PostValidationErrors posts validation errors as a comment to help the user fix format issues.
func (c *Client) PostValidationErrors(issueNumber string, errors []string) error {
	var message strings.Builder
	message.WriteString("## ⚠️ Hour Logging Format Errors\n\n")
	message.WriteString("I found some issues with the hours format:\n\n")

	for i, err := range errors {
		message.WriteString(fmt.Sprintf("%d. %s\n", i+1, err))
	}

	message.WriteString("\n### Correct Format\n")
	message.WriteString("Use comma-separated entries:\n")
	message.WriteString("```\n")
	message.WriteString("Hours: 16: 3h, 17: 4.5h, 18: 2h\n")
	message.WriteString("```\n\n")
	message.WriteString("**Date formats:**\n")
	message.WriteString("- Day only: `16: 3h` (uses current month/year)\n")
	message.WriteString("- Month-Day: `03-16: 4h` (uses current year)\n")
	message.WriteString("- Full date: `2024-02-16: 5h`\n\n")
	message.WriteString("**To delete an entry:** Set hours to 0h\n")
	message.WriteString("```\n")
	message.WriteString("Hours: 16: 0h\n")
	message.WriteString("```\n")

	return c.PostCommentWithBody(issueNumber, message.String())
}

// GetAllLoggedHours retrieves all logged hours from bot commands for an issue.
// This aggregates hours from all @wrikemeup log commands in the issue's comments.
func (c *Client) GetAllLoggedHours(issueNumber string) (map[string]float64, error) {
	metadata, err := c.GetIssueMetadata(issueNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue metadata: %w", err)
	}

	// Return the daily hours from metadata (which includes all parsed comments)
	if metadata.DailyHours != nil && len(metadata.DailyHours) > 0 {
		return metadata.DailyHours, nil
	}

	return make(map[string]float64), nil
}
