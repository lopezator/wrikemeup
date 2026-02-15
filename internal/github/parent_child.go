package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// IssueType represents the type of issue in the parent/child hierarchy.
type IssueType int

const (
	IssueTypeStandalone IssueType = iota // No parent, no children
	IssueTypeChild                        // Has a parent
	IssueTypeParent                       // Has children
)

// IssueRelationship holds parent/child information for an issue.
type IssueRelationship struct {
	Type       IssueType
	ParentNum  int   // Parent issue number (if child)
	ChildNums  []int // Child issue numbers (if parent)
	IssueNum   int   // This issue's number
}

// GetIssueRelationship determines if an issue is standalone, a child, or a parent.
func (c *Client) GetIssueRelationship(issueNumber string) (*IssueRelationship, error) {
	issueNum, err := strconv.Atoi(issueNumber)
	if err != nil {
		return nil, fmt.Errorf("invalid issue number: %s", issueNumber)
	}
	
	// Get the issue body to check for parent references
	metadata, err := c.GetIssueMetadata(issueNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue metadata: %w", err)
	}
	
	// Check if this issue references a parent
	parentNum := findParentReference(metadata.Body, issueNum)
	
	// Search for children that reference this issue
	children, err := c.findChildIssues(issueNum)
	if err != nil {
		return nil, fmt.Errorf("failed to find children: %w", err)
	}
	
	rel := &IssueRelationship{
		IssueNum:  issueNum,
		ParentNum: parentNum,
		ChildNums: children,
	}
	
	// Determine type
	if parentNum > 0 {
		rel.Type = IssueTypeChild
	} else if len(children) > 0 {
		rel.Type = IssueTypeParent
	} else {
		rel.Type = IssueTypeStandalone
	}
	
	return rel, nil
}

// findParentReference looks for parent references in the issue body.
// Patterns:
//   - Parent: #123
//   - Related to #123
//   - Part of #123
//   - Tasklist: - [ ] #123 (reversed - child contains parent)
func findParentReference(body string, issueNum int) int {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)parent:\s*#(\d+)`),
		regexp.MustCompile(`(?i)related\s+to:\s*#(\d+)`),
		regexp.MustCompile(`(?i)part\s+of:\s*#(\d+)`),
	}
	
	for _, pattern := range patterns {
		if matches := pattern.FindStringSubmatch(body); len(matches) >= 2 {
			if parentNum, err := strconv.Atoi(matches[1]); err == nil {
				return parentNum
			}
		}
	}
	
	return 0
}

// findChildIssues searches for issues that reference this issue as parent.
func (c *Client) findChildIssues(issueNum int) ([]int, error) {
	// Use GitHub Search API to find issues that reference this one
	// Search queries:
	// 1. "Parent: #123" in body
	// 2. "Related to #123" in body
	// 3. "Part of #123" in body
	// 4. Tasklist references: - [ ] #123
	
	queries := []string{
		fmt.Sprintf(`"Parent: #%d" in:body repo:%s`, issueNum, c.repo),
		fmt.Sprintf(`"Related to #%d" in:body repo:%s`, issueNum, c.repo),
		fmt.Sprintf(`"Part of #%d" in:body repo:%s`, issueNum, c.repo),
		fmt.Sprintf(`"#%d" in:body is:issue repo:%s`, issueNum, c.repo),
	}
	
	childMap := make(map[int]bool)
	
	for _, query := range queries {
		results, err := c.searchIssues(query)
		if err != nil {
			// Continue with other queries even if one fails
			continue
		}
		
		for _, childNum := range results {
			if childNum != issueNum {
				childMap[childNum] = true
			}
		}
	}
	
	// Convert map to slice
	children := make([]int, 0, len(childMap))
	for childNum := range childMap {
		children = append(children, childNum)
	}
	
	return children, nil
}

// searchIssues performs a GitHub search and returns issue numbers.
func (c *Client) searchIssues(query string) ([]int, error) {
	encodedQuery := url.QueryEscape(query)
	searchURL := fmt.Sprintf("https://api.github.com/search/issues?q=%s&per_page=100", encodedQuery)
	
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create search request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+c.botToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	
	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search API error: %s (status: %d)", string(body), resp.StatusCode)
	}
	
	var result struct {
		Items []struct {
			Number int    `json:"number"`
			Body   string `json:"body"`
		} `json:"items"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode search results: %w", err)
	}
	
	numbers := make([]int, 0, len(result.Items))
	for _, item := range result.Items {
		numbers = append(numbers, item.Number)
	}
	
	return numbers, nil
}

// GetAllChildrenComments retrieves all comments from all child issues.
// This is used for the full scan aggregation algorithm.
func (c *Client) GetAllChildrenComments(children []int) (map[int][]Comment, error) {
	result := make(map[int][]Comment)
	
	for _, childNum := range children {
		comments, err := c.GetIssueComments(strconv.Itoa(childNum))
		if err != nil {
			// Log error but continue with other children
			fmt.Printf("Warning: failed to get comments for child #%d: %v\n", childNum, err)
			continue
		}
		
		result[childNum] = comments
	}
	
	return result, nil
}

// AggregateHoursFromChildren implements the full scan aggregation algorithm.
// For each date in targetDates:
//   - Scan ALL children's comment history
//   - Find latest @wrikemeup log mentioning this date per child
//   - Sum hours across all children for that date
//   - Support 0h deletion (removes child's entry)
func AggregateHoursFromChildren(childComments map[int][]Comment, targetDates []string) (map[string]float64, error) {
	aggregated := make(map[string]float64)
	
	for _, targetDate := range targetDates {
		totalHours := 0.0
		
		// For each child, find latest log for this date
		for _, comments := range childComments {
			latestHours := findLatestHoursForDate(comments, targetDate)
			
			// 0h means delete this child's entry
			if latestHours >= 0 {
				totalHours += latestHours
			}
		}
		
		aggregated[targetDate] = totalHours
	}
	
	return aggregated, nil
}

// findLatestHoursForDate finds the most recent log entry for a specific date.
// Returns -1 if no entry found, 0 if explicitly set to 0h (deletion).
func findLatestHoursForDate(comments []Comment, targetDate string) float64 {
	latestHours := -1.0
	latestTime := ""
	
	for _, comment := range comments {
		// Check if this is a log command
		if !strings.Contains(comment.Body, "@wrikemeup log") {
			continue
		}
		
		// Try to parse as spec format
		entries, err := ParseSpecLogCommand(comment.Body, parseCommentTime(comment.CreatedAt))
		if err != nil {
			// Not a valid log command, skip
			continue
		}
		
		// Check if any entry matches our target date
		for _, entry := range entries {
			if entry.Date == targetDate {
				// This comment mentions our target date
				// Use it if it's newer than what we've seen
				if comment.CreatedAt > latestTime {
					latestTime = comment.CreatedAt
					latestHours = entry.Hours
				}
			}
		}
	}
	
	return latestHours
}

// parseCommentTime parses the GitHub comment timestamp.
func parseCommentTime(timestamp string) time.Time {
	// GitHub timestamps are in RFC3339 format
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		// Fall back to current time if parsing fails
		return time.Now()
	}
	return t
}
