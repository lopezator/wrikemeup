package github

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// LogEntry represents a single time log entry from the specification format.
// Format: <date> <duration> ["description"]
// Example: today 3h "feature work"
type LogEntry struct {
	Date        string  // Resolved to YYYY-MM-DD format
	Hours       float64 // Total hours (converted from h/m/h+m)
	Description string  // Optional description
}

var (
	// Specification format: <date> <duration> ["description"]
	// Examples: today 3h, last monday 4h, feb 15 5h "code review"
	specEntryRegex = regexp.MustCompile(`([a-zA-Z0-9\s-]+?)\s+((?:\d+h)?(?:\d+m)?)(?:\s+"([^"]*)")?`)
)

// ParseSpecLogCommand parses the specification format log command.
// Format: @wrikemeup log <entries>
// Where <entries> is comma-separated: <date> <duration> ["description"]
// Examples:
//   @wrikemeup log today 3h "feature work"
//   @wrikemeup log today 3h, yesterday 2h "code review"
//   @wrikemeup log last monday 4h, feb 15 5h
func ParseSpecLogCommand(comment string, now time.Time) ([]LogEntry, error) {
	comment = strings.TrimSpace(comment)
	
	// Extract the log command content
	if !strings.HasPrefix(comment, "@wrikemeup log ") {
		return nil, errors.New("not a log command")
	}
	
	content := strings.TrimPrefix(comment, "@wrikemeup log ")
	content = strings.TrimSpace(content)
	
	if content == "" {
		return nil, errors.New("log command requires entries")
	}
	
	// Split by comma to get individual entries
	rawEntries := splitEntries(content)
	
	var entries []LogEntry
	for _, raw := range rawEntries {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		
		entry, err := parseSpecEntry(raw, now)
		if err != nil {
			return nil, fmt.Errorf("invalid entry '%s': %w", raw, err)
		}
		
		entries = append(entries, entry)
	}
	
	if len(entries) == 0 {
		return nil, errors.New("no valid entries found")
	}
	
	return entries, nil
}

// splitEntries splits comma-separated entries, handling quoted descriptions.
func splitEntries(content string) []string {
	var entries []string
	var current strings.Builder
	inQuotes := false
	
	for i := 0; i < len(content); i++ {
		ch := content[i]
		
		if ch == '"' {
			inQuotes = !inQuotes
			current.WriteByte(ch)
		} else if ch == ',' && !inQuotes {
			entries = append(entries, current.String())
			current.Reset()
		} else {
			current.WriteByte(ch)
		}
	}
	
	// Add the last entry
	if current.Len() > 0 {
		entries = append(entries, current.String())
	}
	
	return entries
}

// parseSpecEntry parses a single entry in specification format.
// Format: <date> <duration> ["description"]
// Examples: today 3h, last monday 4h, feb 15 5h "code review"
func parseSpecEntry(entry string, now time.Time) (LogEntry, error) {
	entry = strings.TrimSpace(entry)
	
	// Split into parts: date duration [description]
	// We need to be careful about dates with spaces like "last monday" or "feb 15"
	parts := strings.Fields(entry)
	if len(parts) < 2 {
		return LogEntry{}, errors.New("entry must have at least date and duration")
	}
	
	// Find the duration part (contains 'h' or 'm')
	durationIdx := -1
	for i, part := range parts {
		if strings.Contains(part, "h") || strings.Contains(part, "m") {
			durationIdx = i
			break
		}
	}
	
	if durationIdx == -1 {
		return LogEntry{}, errors.New("no duration found (must contain 'h' or 'm')")
	}
	
	// Everything before duration is the date
	dateStr := strings.Join(parts[0:durationIdx], " ")
	durationStr := parts[durationIdx]
	
	// Everything after duration is the description (if any)
	var description string
	if durationIdx+1 < len(parts) {
		// Join remaining parts and strip quotes
		description = strings.Join(parts[durationIdx+1:], " ")
		description = strings.Trim(description, `"`)
	}
	
	// Parse date
	date, err := ParseSpecDate(dateStr, now)
	if err != nil {
		return LogEntry{}, fmt.Errorf("invalid date '%s': %w", dateStr, err)
	}
	
	// Parse duration
	hours, err := ParseDuration(durationStr)
	if err != nil {
		return LogEntry{}, fmt.Errorf("invalid duration '%s': %w", durationStr, err)
	}
	
	return LogEntry{
		Date:        date,
		Hours:       hours,
		Description: description,
	}, nil
}

// ParseSpecDate parses date formats from the specification.
// Supports:
//   - Relative: today, yesterday
//   - This/Next Week: monday, tuesday, etc.
//   - Last Week: last monday, last tuesday, etc.
//   - Specific: 2026-02-15, feb 15, 15 feb
func ParseSpecDate(dateStr string, now time.Time) (string, error) {
	dateStr = strings.ToLower(strings.TrimSpace(dateStr))
	
	// Handle "today"
	if dateStr == "today" {
		return now.Format("2006-01-02"), nil
	}
	
	// Handle "yesterday"
	if dateStr == "yesterday" {
		return now.AddDate(0, 0, -1).Format("2006-01-02"), nil
	}
	
	// Handle "last <weekday>"
	if strings.HasPrefix(dateStr, "last ") {
		weekdayStr := strings.TrimPrefix(dateStr, "last ")
		return parseLastWeekday(weekdayStr, now)
	}
	
	// Handle weekday names (this week or last week)
	if weekday := parseWeekday(dateStr); weekday != -1 {
		return findRecentWeekday(weekday, now), nil
	}
	
	// Handle ISO format: 2026-02-15
	if matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, dateStr); matched {
		return dateStr, nil
	}
	
	// Handle text month formats: "feb 15", "15 feb", "march 20", "20 march"
	if date := parseTextMonthDate(dateStr, now); date != "" {
		return date, nil
	}
	
	// Handle numeric formats: "15" (day only), "02-15" (month-day)
	if matched, _ := regexp.MatchString(`^\d{1,2}$`, dateStr); matched {
		day, _ := strconv.Atoi(dateStr)
		return fmt.Sprintf("%d-%02d-%02d", now.Year(), now.Month(), day), nil
	}
	
	if matched, _ := regexp.MatchString(`^\d{2}-\d{2}$`, dateStr); matched {
		return fmt.Sprintf("%d-%s", now.Year(), dateStr), nil
	}
	
	return "", fmt.Errorf("unrecognized date format: %s", dateStr)
}

// ParseDuration parses duration from specification formats.
// Supports:
//   - Hours: 3h, 4.5h
//   - Minutes: 30m
//   - Combined: 2h30m
func ParseDuration(durationStr string) (float64, error) {
	durationStr = strings.TrimSpace(strings.ToLower(durationStr))
	
	// Combined format: 2h30m
	combinedRegex := regexp.MustCompile(`^(\d+)h(\d+)m$`)
	if matches := combinedRegex.FindStringSubmatch(durationStr); len(matches) == 3 {
		hours, _ := strconv.ParseFloat(matches[1], 64)
		minutes, _ := strconv.ParseFloat(matches[2], 64)
		return hours + (minutes / 60.0), nil
	}
	
	// Hours only: 3h, 4.5h
	if strings.HasSuffix(durationStr, "h") {
		hoursStr := strings.TrimSuffix(durationStr, "h")
		hours, err := strconv.ParseFloat(hoursStr, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid hours: %s", hoursStr)
		}
		return hours, nil
	}
	
	// Minutes only: 30m
	if strings.HasSuffix(durationStr, "m") {
		minutesStr := strings.TrimSuffix(durationStr, "m")
		minutes, err := strconv.ParseFloat(minutesStr, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid minutes: %s", minutesStr)
		}
		return minutes / 60.0, nil
	}
	
	return 0, fmt.Errorf("invalid duration format (must end with 'h' or 'm'): %s", durationStr)
}

// parseLastWeekday finds the date for "last <weekday>".
func parseLastWeekday(weekdayStr string, now time.Time) (string, error) {
	weekday := parseWeekday(weekdayStr)
	if weekday == -1 {
		return "", fmt.Errorf("invalid weekday: %s", weekdayStr)
	}
	
	// Go back to last week's instance of this weekday
	currentWeekday := now.Weekday()
	daysBack := int(currentWeekday - weekday)
	
	// If target is today or in the future this week, go back a full week
	if daysBack <= 0 {
		daysBack += 7
	}
	
	targetDate := now.AddDate(0, 0, -daysBack)
	return targetDate.Format("2006-01-02"), nil
}

// findRecentWeekday finds the most recent occurrence of a weekday.
func findRecentWeekday(weekday time.Weekday, now time.Time) string {
	currentWeekday := now.Weekday()
	daysBack := int(currentWeekday - weekday)
	
	// If target is in the future this week, go back to last week
	if daysBack < 0 {
		daysBack += 7
	}
	
	// If it's today, use today
	if daysBack == 0 {
		return now.Format("2006-01-02")
	}
	
	targetDate := now.AddDate(0, 0, -daysBack)
	return targetDate.Format("2006-01-02")
}

// parseWeekday converts weekday name to time.Weekday.
func parseWeekday(name string) time.Weekday {
	weekdays := map[string]time.Weekday{
		"monday": time.Monday, "mon": time.Monday,
		"tuesday": time.Tuesday, "tue": time.Tuesday,
		"wednesday": time.Wednesday, "wed": time.Wednesday,
		"thursday": time.Thursday, "thu": time.Thursday,
		"friday": time.Friday, "fri": time.Friday,
		"saturday": time.Saturday, "sat": time.Saturday,
		"sunday": time.Sunday, "sun": time.Sunday,
	}
	
	if wd, ok := weekdays[name]; ok {
		return wd
	}
	return -1
}

// parseTextMonthDate handles formats like "feb 15", "15 feb", "march 20", "20 march".
func parseTextMonthDate(dateStr string, now time.Time) string {
	months := map[string]time.Month{
		"jan": time.January, "january": time.January,
		"feb": time.February, "february": time.February,
		"mar": time.March, "march": time.March,
		"apr": time.April, "april": time.April,
		"may": time.May,
		"jun": time.June, "june": time.June,
		"jul": time.July, "july": time.July,
		"aug": time.August, "august": time.August,
		"sep": time.September, "september": time.September,
		"oct": time.October, "october": time.October,
		"nov": time.November, "november": time.November,
		"dec": time.December, "december": time.December,
	}
	
	parts := strings.Fields(dateStr)
	if len(parts) != 2 {
		return ""
	}
	
	// Try "month day" format: "feb 15"
	if month, ok := months[parts[0]]; ok {
		if day, err := strconv.Atoi(parts[1]); err == nil && day >= 1 && day <= 31 {
			return fmt.Sprintf("%d-%02d-%02d", now.Year(), month, day)
		}
	}
	
	// Try "day month" format: "15 feb"
	if month, ok := months[parts[1]]; ok {
		if day, err := strconv.Atoi(parts[0]); err == nil && day >= 1 && day <= 31 {
			return fmt.Sprintf("%d-%02d-%02d", now.Year(), month, day)
		}
	}
	
	return ""
}
