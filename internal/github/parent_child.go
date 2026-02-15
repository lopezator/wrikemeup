package github

import (
	"fmt"
	"regexp"
	"strconv"
)

// IssueType represents the type of issue in parent/child hierarchy.
type IssueType string

const (
	IssueTypeStandalone IssueType = "standalone" // No parent, no children
	IssueTypeChild      IssueType = "child"      // Has parent
	IssueTypeParent     IssueType = "parent"     // Has children
)

// IssueRelationship contains parent/child relationship information.
type IssueRelationship struct {
	Type        IssueType
	ParentNum   int   // Parent issue number (for children)
	ChildrenNum []int // Child issue numbers (for parents)
}

// GetIssueRelationship determines if an issue is standalone, child, or parent.
func (c *Client) GetIssueRelationship(issueNumber string) (*IssueRelationship, error) {
	issueNum, err := strconv.Atoi(issueNumber)
	if err != nil {
		return nil, fmt.Errorf("invalid issue number: %w", err)
	}

	// Get issue metadata to check for parent reference
	metadata, err := c.GetIssueMetadata(issueNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue metadata: %w", err)
	}

	// Check if this issue has a parent reference in body
	parentNum := extractParentReference(metadata.Body)

	// Find all children that reference this issue
	children, err := c.GetChildIssues(issueNum)
	if err != nil {
		return nil, fmt.Errorf("failed to get child issues: %w", err)
	}

	// Determine type
	rel := &IssueRelationship{
		ParentNum:   parentNum,
		ChildrenNum: children,
	}

	if parentNum > 0 {
		rel.Type = IssueTypeChild
	} else if len(children) > 0 {
		rel.Type = IssueTypeParent
	} else {
		rel.Type = IssueTypeStandalone
	}

	return rel, nil
}

// extractParentReference extracts parent issue number from issue body.
// Looks for patterns: "Parent: #123", "Related to #123", "Part of #123"
func extractParentReference(body string) int {
	patterns := []string{
		`(?i)parent:\s*#(\d+)`,
		`(?i)related to\s*#(\d+)`,
		`(?i)part of\s*#(\d+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(body); len(matches) > 1 {
			if num, err := strconv.Atoi(matches[1]); err == nil {
				return num
			}
		}
	}

	return 0
}

// AggregateHoursFromChildren implements the "Full Scan with Date Filter" algorithm from spec.
// For each target date:
// 1. Get all children
// 2. Scan ALL children's comment history
// 3. Find latest @wrikemeup log per child for that date
// 4. Sum hours across children
func (c *Client) AggregateHoursFromChildren(parentIssueNum int, targetDates []string) (map[string]float64, error) {
	// Get all children
	children, err := c.GetChildIssues(parentIssueNum)
	if err != nil {
		return nil, fmt.Errorf("failed to get children: %w", err)
	}

	if len(children) == 0 {
		return make(map[string]float64), nil
	}

	// For each child, get all comments
	childComments := make(map[int][]Comment)
	for _, childNum := range children {
		comments, err := c.GetIssueComments(strconv.Itoa(childNum))
		if err != nil {
			return nil, fmt.Errorf("failed to get comments for child #%d: %w", childNum, err)
		}
		childComments[childNum] = comments
	}

	// Aggregate hours per date
	aggregated := make(map[string]float64)

	for _, targetDate := range targetDates {
		totalForDate := 0.0

		// For each child, find latest log containing this date
		for childNum, comments := range childComments {
			hours := findLatestHoursForDate(comments, targetDate)
			if hours > 0 {
				totalForDate += hours
			} else if hours == 0 {
				// 0h means delete this child's entry (per spec)
				// Don't add to total
			}
			_ = childNum // For debugging
		}

		aggregated[targetDate] = totalForDate
	}

	return aggregated, nil
}

// findLatestHoursForDate scans all comments to find the latest @wrikemeup log mentioning this date.
// Returns hours for that date (0h means delete).
func findLatestHoursForDate(comments []Comment, targetDate string) float64 {
	var latestHours float64 = -1 // -1 means not found

	// Scan from oldest to newest (newer ones will overwrite)
	for _, comment := range comments {
		// Try to parse as log command
		entries, err := ParseSpecLog(comment.Body)
		if err != nil {
			continue // Not a log command
		}

		// Check if this log mentions the target date
		for _, entry := range entries {
			if entry.Date == targetDate {
				latestHours = entry.Hours
				// Don't break - keep scanning for even later logs
			}
		}
	}

	if latestHours == -1 {
		return -1 // Not found
	}
	return latestHours
}

// GetAllChildrenComments gets all comments from all children (for full scan).
func (c *Client) GetAllChildrenComments(parentIssueNum int) (map[int][]Comment, error) {
	children, err := c.GetChildIssues(parentIssueNum)
	if err != nil {
		return nil, fmt.Errorf("failed to get children: %w", err)
	}

	result := make(map[int][]Comment)
	for _, childNum := range children {
		comments, err := c.GetIssueComments(strconv.Itoa(childNum))
		if err != nil {
			return nil, fmt.Errorf("failed to get comments for child #%d: %w", childNum, err)
		}
		result[childNum] = comments
	}

	return result, nil
}

// GetTargetWrikeTask determines which issue should have the Wrike task.
// Returns issue number and whether it needs to be created.
func (c *Client) GetTargetWrikeTask(issueNumber string) (targetIssue string, needsCreate bool, err error) {
	rel, err := c.GetIssueRelationship(issueNumber)
	if err != nil {
		return "", false, err
	}

	var targetNum int

	switch rel.Type {
	case IssueTypeStandalone:
		// Use own issue
		targetNum, _ = strconv.Atoi(issueNumber)

	case IssueTypeChild:
		// Use parent's issue
		targetNum = rel.ParentNum

	case IssueTypeParent:
		// Use own issue
		targetNum, _ = strconv.Atoi(issueNumber)
	}

	targetIssue = strconv.Itoa(targetNum)

	// Check if target has Wrike ID
	metadata, err := c.GetIssueMetadata(targetIssue)
	if err != nil {
		return "", false, fmt.Errorf("failed to get target metadata: %w", err)
	}

	needsCreate = (metadata.WrikeTaskID == "")

	return targetIssue, needsCreate, nil
}
