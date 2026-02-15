package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

// IssueMetadata holds metadata about a GitHub issue.
type IssueMetadata struct {
	Number          int
	Title           string
	Body            string
	WrikeTaskID     string
	Hours           float64
	DailyHours      map[string]float64 // Date -> Hours mapping
	LastSyncedHours float64            // Track last synced amount for incremental logging
	SubIssues       []int
}

// Comment represents a GitHub issue comment.
type Comment struct {
	ID        int64  `json:"id"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

var (
	hoursRegex       = regexp.MustCompile(`(?i)hours?:\s*([\d.]+)h?`)
	dailyHoursRegex  = regexp.MustCompile(`(?i)hours?:\s*(\d{4}-\d{2}-\d{2}):\s*([\d.]+)h?`)
	lastSyncedRegex  = regexp.MustCompile(`(?i)last\s+synced:\s*([\d.]+)h`)
	wrikeTaskRegex   = regexp.MustCompile(`(?i)wrike\s*task\s*id?:\s*([A-Za-z0-9_-]+)`)
	subIssuesRegex   = regexp.MustCompile(`#(\d+)`)
	parentRefRegex   = regexp.MustCompile(`(?i)(parent|related to|part of)[:\s]*#%d`)
	tasklistRefRegex = regexp.MustCompile(`-\s*\[[ x]\]\s*#%d`)
	issueRefRegex    = regexp.MustCompile(`\b#%d\b`)
	hoursBlockRegex  = regexp.MustCompile(`(?i)hours?:\s*\n((?:\s*-\s*.+\n?)+)`)
	hoursEntryRegex  = regexp.MustCompile(`-\s*(\d{4}-\d{2}-\d{2}|\d{2}-\d{2}|\d{1,2})\s*=\s*([\d.]+)h?`)
)

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

	// Initialize daily hours map
	metadata.DailyHours = make(map[string]float64)

	// Get current date for smart date parsing
	currentDate := time.Now().Format("2006-01-02")

	// Fetch and parse hours from comments (new format)
	comments, err := c.GetIssueComments(issueNumber)
	if err != nil {
		log.Printf("Warning: failed to get comments for issue #%s: %v", issueNumber, err)
	} else {
		commentHours, err := ParseHoursFromComments(comments, currentDate)
		if err != nil {
			log.Printf("Warning: failed to parse hours from comments: %v", err)
		} else {
			// Add hours from comments to daily hours
			for date, hours := range commentHours {
				metadata.DailyHours[date] += hours
				metadata.Hours += hours
			}
		}
	}

	// Also parse old format from issue body for backward compatibility
	dailyMatches := dailyHoursRegex.FindAllStringSubmatch(issue.Body, -1)
	for _, match := range dailyMatches {
		if len(match) >= 3 {
			date := match[1]
			if hours, err := strconv.ParseFloat(match[2], 64); err == nil {
				metadata.DailyHours[date] += hours
				metadata.Hours += hours
			}
		}
	}

	// Parse simple hours format if no daily breakdown found (e.g., "Hours: 4.5h")
	if len(metadata.DailyHours) == 0 {
		if matches := hoursRegex.FindStringSubmatch(issue.Body); len(matches) >= 2 {
			if hours, err := strconv.ParseFloat(matches[1], 64); err == nil {
				metadata.Hours = hours
			}
		}
	}

	// Parse last synced hours for incremental tracking
	if matches := lastSyncedRegex.FindStringSubmatch(issue.Body); len(matches) >= 2 {
		if hours, err := strconv.ParseFloat(matches[1], 64); err == nil {
			metadata.LastSyncedHours = hours
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

// UpdateLastSyncedHours updates the "Last Synced" marker in the issue body.
func (c *Client) UpdateLastSyncedHours(issueNumber string, totalHours float64) error {
	metadata, err := c.GetIssueMetadata(issueNumber)
	if err != nil {
		return err
	}

	body := metadata.Body
	syncLine := fmt.Sprintf("Last Synced: %.2fh", totalHours)

	// If Last Synced already exists, update it
	if lastSyncedRegex.MatchString(body) {
		body = lastSyncedRegex.ReplaceAllString(body, syncLine)
	} else {
		// Add it after the Wrike Task ID line or at the beginning
		if metadata.WrikeTaskID != "" {
			// Add after Wrike Task ID
			wrikeTaskLine := fmt.Sprintf("Wrike Task ID: %s", metadata.WrikeTaskID)
			body = regexp.MustCompile(regexp.QuoteMeta(wrikeTaskLine)).ReplaceAllString(body, wrikeTaskLine+"\n"+syncLine)
		} else {
			// Add at the beginning
			if body != "" {
				body = syncLine + "\n\n" + body
			} else {
				body = syncLine
			}
		}
	}

	return c.UpdateIssueBody(issueNumber, body)
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

// GetChildIssues retrieves all child issues (sub-issues) for a parent issue.
// It searches for issues that have "Parent: #<issueNumber>" or are in a tasklist.
func (c *Client) GetChildIssues(issueNumber int) ([]int, error) {
	// Search for issues that reference this issue as parent
	// GitHub's issue search supports finding issues that link to others
	owner, repo := c.splitRepo()

	// Search for issues in the same repo that might be children
	// We'll look for issues that mention this issue number
	searchQuery := fmt.Sprintf("repo:%s/%s %d in:body state:open,closed", owner, repo, issueNumber)

	urlStr := fmt.Sprintf("https://api.github.com/search/issues?q=%s&per_page=100", url.QueryEscape(searchQuery))
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("github: failed to create search request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.botToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github: search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github: search API error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var searchResult struct {
		Items []struct {
			Number int    `json:"number"`
			Body   string `json:"body"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("github: failed to decode search response: %w", err)
	}

	var childIssues []int

	// Compile patterns with specific issue number (cannot be pre-compiled at package level)
	parentPattern := regexp.MustCompile(fmt.Sprintf(`(?i)(parent|related to|part of)[:\s]*#%d`, issueNumber))
	tasklistPattern := regexp.MustCompile(fmt.Sprintf(`-\s*\[[ x]\]\s*#%d`, issueNumber))

	for _, item := range searchResult.Items {
		// Skip the parent issue itself
		if item.Number == issueNumber {
			continue
		}

		// Check if this issue references the parent using pre-compiled patterns
		if parentPattern.MatchString(item.Body) || tasklistPattern.MatchString(item.Body) {
			childIssues = append(childIssues, item.Number)
		}
	}

	return childIssues, nil
}

// splitRepo splits the repo string into owner and name.
func (c *Client) splitRepo() (string, string) {
	parts := bytes.Split([]byte(c.repo), []byte("/"))
	if len(parts) != 2 {
		return "", ""
	}
	return string(parts[0]), string(parts[1])
}

// GetIssueComments retrieves all comments for a GitHub issue.
func (c *Client) GetIssueComments(issueNumber string) ([]Comment, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/issues/%s/comments?per_page=100", c.repo, issueNumber)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("github: failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.botToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github: failed to get comments: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github: API error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var comments []Comment
	if err := json.NewDecoder(resp.Body).Decode(&comments); err != nil {
		return nil, fmt.Errorf("github: failed to decode response: %w", err)
	}

	return comments, nil
}

// ParseSmartDate converts smart date formats to full YYYY-MM-DD format.
// Supports: "16" (day), "03-16" (month-day), "2023-03-16" (full date)
func ParseSmartDate(dateStr string, currentDate string) (string, error) {
	// currentDate should be in YYYY-MM-DD format
	parts := regexp.MustCompile(`-`).Split(currentDate, -1)
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid current date format: %s", currentDate)
	}
	currentYear, currentMonth := parts[0], parts[1]

	// Check format
	if regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`).MatchString(dateStr) {
		// Full date: 2023-03-16
		return dateStr, nil
	} else if regexp.MustCompile(`^\d{2}-\d{2}$`).MatchString(dateStr) {
		// Month-day: 03-16 (use current year)
		return fmt.Sprintf("%s-%s", currentYear, dateStr), nil
	} else if regexp.MustCompile(`^\d{1,2}$`).MatchString(dateStr) {
		// Day only: 16 (use current year and month)
		day := dateStr
		if len(day) == 1 {
			day = "0" + day
		}
		return fmt.Sprintf("%s-%s-%s", currentYear, currentMonth, day), nil
	}

	return "", fmt.Errorf("invalid date format: %s", dateStr)
}

// ParseHoursFromComments extracts hours from comments using the new format.
// Format: Hours:\n- 16 = 3h\n- 03-17 = 4h\n- 2023-12-25 = 5h
func ParseHoursFromComments(comments []Comment, currentDate string) (map[string]float64, error) {
	dailyHours := make(map[string]float64)

	for _, comment := range comments {
		// Find Hours: blocks in comment
		blockMatches := hoursBlockRegex.FindAllStringSubmatch(comment.Body, -1)
		for _, blockMatch := range blockMatches {
			if len(blockMatch) < 2 {
				continue
			}
			block := blockMatch[1]

			// Parse each entry in the block
			entryMatches := hoursEntryRegex.FindAllStringSubmatch(block, -1)
			for _, entryMatch := range entryMatches {
				if len(entryMatch) < 3 {
					continue
				}

				dateStr := entryMatch[1]
				hoursStr := entryMatch[2]

				// Parse smart date
				fullDate, err := ParseSmartDate(dateStr, currentDate)
				if err != nil {
					log.Printf("Warning: failed to parse date '%s': %v", dateStr, err)
					continue
				}

				// Parse hours
				hours, err := strconv.ParseFloat(hoursStr, 64)
				if err != nil {
					log.Printf("Warning: failed to parse hours '%s': %v", hoursStr, err)
					continue
				}

				// Add to daily hours (accumulate if same date)
				dailyHours[fullDate] += hours
			}
		}
	}

	return dailyHours, nil
}
