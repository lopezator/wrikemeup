package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
)

// IssueMetadata holds metadata about a GitHub issue.
type IssueMetadata struct {
	Number      int
	Title       string
	Body        string
	WrikeTaskID string
	Hours       float64
	SubIssues   []int
}

var hoursRegex = regexp.MustCompile(`(?i)hours?:\s*([\d.]+)h?`)
var wrikeTaskRegex = regexp.MustCompile(`(?i)wrike\s*task\s*id?:\s*([A-Za-z0-9_-]+)`)
var subIssuesRegex = regexp.MustCompile(`#(\d+)`)

// GetIssueMetadata retrieves metadata for a GitHub issue.
func (c *Client) GetIssueMetadata(issueNumber string) (*IssueMetadata, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/issues/%s", c.repo, issueNumber)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("github: failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.botToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github: failed to get issue: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github: API error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var issue struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		Body   string `json:"body"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("github: failed to decode response: %w", err)
	}

	metadata := &IssueMetadata{
		Number: issue.Number,
		Title:  issue.Title,
		Body:   issue.Body,
	}

	// Parse Wrike task ID from body
	if matches := wrikeTaskRegex.FindStringSubmatch(issue.Body); len(matches) >= 2 {
		metadata.WrikeTaskID = matches[1]
	}

	// Parse hours from body
	if matches := hoursRegex.FindStringSubmatch(issue.Body); len(matches) >= 2 {
		if hours, err := strconv.ParseFloat(matches[1], 64); err == nil {
			metadata.Hours = hours
		}
	}

	// Parse sub-issues (referenced as #123)
	subIssueMatches := subIssuesRegex.FindAllStringSubmatch(issue.Body, -1)
	for _, match := range subIssueMatches {
		if len(match) >= 2 {
			if num, err := strconv.Atoi(match[1]); err == nil {
				metadata.SubIssues = append(metadata.SubIssues, num)
			}
		}
	}

	return metadata, nil
}

// UpdateIssueBody updates the body of a GitHub issue.
func (c *Client) UpdateIssueBody(issueNumber string, newBody string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/issues/%s", c.repo, issueNumber)

	payload := map[string]string{
		"body": newBody,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("github: failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("github: failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.botToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("github: failed to update issue: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("github: API error: %s (status: %d)", string(respBody), resp.StatusCode)
	}

	return nil
}

// AddOrUpdateWrikeTaskID adds or updates the Wrike Task ID in the issue body.
func (c *Client) AddOrUpdateWrikeTaskID(issueNumber string, taskID string) error {
	metadata, err := c.GetIssueMetadata(issueNumber)
	if err != nil {
		return err
	}

	body := metadata.Body
	wrikeTaskLine := fmt.Sprintf("Wrike Task ID: %s", taskID)

	// If Wrike Task ID already exists, update it
	if metadata.WrikeTaskID != "" {
		body = wrikeTaskRegex.ReplaceAllString(body, wrikeTaskLine)
	} else {
		// Add it to the beginning of the body
		if body != "" {
			body = wrikeTaskLine + "\n\n" + body
		} else {
			body = wrikeTaskLine
		}
	}

	return c.UpdateIssueBody(issueNumber, body)
}
