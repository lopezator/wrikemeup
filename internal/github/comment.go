package github

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Now returns the current time. Extracted as a variable for testing.
var Now = time.Now

// Command represents a parsed command from a GitHub comment.
type Command struct {
	Action      string             // "log", "link", "loghours", "sync", "delete", "show"
	TaskID      string             // Wrike task ID (for legacy commands)
	Hours       float64            // Hours to log (for legacy loghours)
	DailyHours  map[string]float64 // Date -> Hours mapping for new log command
	DeleteDates []string           // Dates to delete
	IssueNumber string             // GitHub issue number for linking
}

// regexp patterns for different command formats.
var (
	reLegacyLog      = regexp.MustCompile(`@wrikemeup log ([A-Za-z0-9_-]+)`)
	reLink           = regexp.MustCompile(`@wrikemeup link ([A-Za-z0-9_-]+)`)
	reLegacyLogHours = regexp.MustCompile(`@wrikemeup loghours ([A-Za-z0-9_-]+)\s+([\d.]+)h?`)
	reSync           = regexp.MustCompile(`@wrikemeup sync`)
	// New command patterns
	reNewLog = regexp.MustCompile(`@wrikemeup log\s+(.+)`)
	reDelete = regexp.MustCompile(`@wrikemeup delete\s+(.+)`)
	reShow   = regexp.MustCompile(`@wrikemeup show`)
	// Entry pattern: [date:]hours (e.g., "3h", "mon:3h", "15:3h", "2024-03-15:3h")
	reLogEntry = regexp.MustCompile(`(?:([a-zA-Z0-9-]+):)?([\d.]+)h?`)
)

// ParseComment parses the GitHub comment to extract the task ID (legacy function).
func ParseComment(comment string) (string, error) {
	cmd, err := ParseCommand(comment)
	if err != nil {
		return "", err
	}
	if cmd.TaskID == "" {
		return "", errors.New("github: Task ID not found in comment")
	}
	return cmd.TaskID, nil
}

const (
	helpMessage = `Commands:
- @wrikemeup log 3h                    (log 3h today)
- @wrikemeup log 3h, mon:2h            (log 3h today, 2h Monday)
- @wrikemeup log yesterday:5h          (log 5h yesterday)
- @wrikemeup delete mon                (delete Monday's hours)
- @wrikemeup show                      (show logged hours)
- @wrikemeup link <task-id>            (link to Wrike task)
- @wrikemeup sync                      (sync hours now)`
)

// ParseCommand parses the GitHub comment and returns a Command struct.
func ParseCommand(comment string) (*Command, error) {
	comment = strings.TrimSpace(comment)

	// Try to match show command
	if reShow.MatchString(comment) {
		return &Command{
			Action: "show",
		}, nil
	}

	// Try to match sync command
	if reSync.MatchString(comment) {
		return &Command{
			Action: "sync",
		}, nil
	}

	// Try to match delete command
	if matches := reDelete.FindStringSubmatch(comment); len(matches) >= 2 {
		dateSpecs := strings.Split(matches[1], ",")
		deleteDates := make([]string, 0)

		for _, spec := range dateSpecs {
			spec = strings.TrimSpace(spec)
			if spec == "" {
				continue
			}
			date, err := ResolveRelativeDate(spec, time.Now())
			if err != nil {
				return nil, fmt.Errorf("invalid date '%s': %v", spec, err)
			}
			deleteDates = append(deleteDates, date)
		}

		return &Command{
			Action:      "delete",
			DeleteDates: deleteDates,
		}, nil
	}

	// Try to match new log command (with relative dates)
	if matches := reNewLog.FindStringSubmatch(comment); len(matches) >= 2 {
		// Check if this looks like a legacy log command (just task ID)
		if reLegacyLog.MatchString(comment) {
			// Legacy format: @wrikemeup log <task-id>
			taskMatches := reLegacyLog.FindStringSubmatch(comment)
			return &Command{
				Action: "log",
				TaskID: taskMatches[1],
			}, nil
		}

		// New format: @wrikemeup log <entries>
		entries := matches[1]
		dailyHours, err := ParseLogEntries(entries, time.Now())
		if err != nil {
			return nil, fmt.Errorf("invalid log format: %v\n\n%s", err, helpMessage)
		}

		return &Command{
			Action:     "log",
			DailyHours: dailyHours,
		}, nil
	}

	// Try to match legacy loghours command
	if matches := reLegacyLogHours.FindStringSubmatch(comment); len(matches) >= 3 {
		hours, err := strconv.ParseFloat(matches[2], 64)
		if err != nil {
			return nil, errors.New("github: invalid hours format. Use format like '4' or '4.25'")
		}
		return &Command{
			Action: "loghours",
			TaskID: matches[1],
			Hours:  hours,
		}, nil
	}

	// Try to match link command
	if matches := reLink.FindStringSubmatch(comment); len(matches) >= 2 {
		return &Command{
			Action: "link",
			TaskID: matches[1],
		}, nil
	}

	return nil, errors.New("github: command not recognized.\n\n" + helpMessage)
}

// ParseLogEntries parses log entries like "3h, mon:2h, yesterday:5h"
func ParseLogEntries(entries string, now time.Time) (map[string]float64, error) {
	dailyHours := make(map[string]float64)

	// Split by comma
	parts := strings.Split(entries, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Match entry pattern
		matches := reLogEntry.FindStringSubmatch(part)
		if matches == nil || len(matches) < 3 {
			return nil, fmt.Errorf("invalid entry format: '%s'. Expected: '3h' or 'mon:3h'", part)
		}

		dateSpec := matches[1] // Empty for today, or "mon", "yesterday", "15", etc.
		hoursStr := matches[2]

		// Parse hours
		hours, err := strconv.ParseFloat(hoursStr, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid hours '%s': must be a number", hoursStr)
		}

		// Resolve date
		var date string
		if dateSpec == "" {
			// No date specified = today
			date = now.Format("2006-01-02")
		} else {
			// Resolve relative or absolute date
			resolvedDate, err := ResolveRelativeDate(dateSpec, now)
			if err != nil {
				return nil, fmt.Errorf("invalid date '%s': %v", dateSpec, err)
			}
			date = resolvedDate
		}

		// Add/update hours (latest wins)
		dailyHours[date] = hours
	}

	return dailyHours, nil
}

// ResolveRelativeDate converts relative date specs to YYYY-MM-DD format.
// Supports:
// - "mon", "tue", "wed", "thu", "fri", "sat", "sun" (this week or last week)
// - "yesterday", "-1" (yesterday), "-2" (2 days ago), etc.
// - "15" (day 15 of current month)
// - "03-15" (March 15 of current year)
// - "2024-03-15" (specific date)
func ResolveRelativeDate(spec string, now time.Time) (string, error) {
	spec = strings.ToLower(strings.TrimSpace(spec))

	// Check for full date format
	if regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`).MatchString(spec) {
		return spec, nil
	}

	// Check for month-day format
	if regexp.MustCompile(`^\d{2}-\d{2}$`).MatchString(spec) {
		return fmt.Sprintf("%d-%s", now.Year(), spec), nil
	}

	// Check for day only format
	if regexp.MustCompile(`^\d{1,2}$`).MatchString(spec) {
		day := spec
		if len(day) == 1 {
			day = "0" + day
		}
		return fmt.Sprintf("%d-%02d-%s", now.Year(), now.Month(), day), nil
	}

	// Check for "yesterday"
	if spec == "yesterday" {
		return now.AddDate(0, 0, -1).Format("2006-01-02"), nil
	}

	// Check for relative days: "-1", "-2", etc.
	if strings.HasPrefix(spec, "-") {
		daysAgo, err := strconv.Atoi(spec)
		if err != nil {
			return "", fmt.Errorf("invalid relative day: %s", spec)
		}
		return now.AddDate(0, 0, daysAgo).Format("2006-01-02"), nil
	}

	// Check for day of week
	weekdays := map[string]time.Weekday{
		"mon": time.Monday, "monday": time.Monday,
		"tue": time.Tuesday, "tuesday": time.Tuesday,
		"wed": time.Wednesday, "wednesday": time.Wednesday,
		"thu": time.Thursday, "thursday": time.Thursday,
		"fri": time.Friday, "friday": time.Friday,
		"sat": time.Saturday, "saturday": time.Saturday,
		"sun": time.Sunday, "sunday": time.Sunday,
	}

	if targetWeekday, ok := weekdays[spec]; ok {
		// Find the most recent occurrence of this weekday
		currentWeekday := now.Weekday()
		daysBack := int(currentWeekday - targetWeekday)

		// If target is in the future this week, go back to last week
		if daysBack < 0 {
			daysBack += 7
		}

		// If it's today, use today
		if daysBack == 0 {
			return now.Format("2006-01-02"), nil
		}

		targetDate := now.AddDate(0, 0, -daysBack)
		return targetDate.Format("2006-01-02"), nil
	}

	return "", fmt.Errorf("unknown date format: %s", spec)
}
