package github

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// SpecEntry represents a single parsed entry from the specification format.
// Format: <date> <duration> ["description"]
// Example: today 3h "fixed bug"
type SpecEntry struct {
	Date        string  // Resolved to YYYY-MM-DD
	Hours       float64 // Total hours
	Description string  // Optional description
}

// ParseSpecLog parses the specification format log command.
// Format: @wrikemeup log <entries>
// Where entries: <date> <duration> ["description"], ...
// Examples:
//   @wrikemeup log today 3h
//   @wrikemeup log today 3h "fixed bug"
//   @wrikemeup log today 3h, yesterday 2h "code review"
func ParseSpecLog(comment string) ([]SpecEntry, error) {
	comment = strings.TrimSpace(comment)
	
	// Check for log command
	if !strings.HasPrefix(comment, "@wrikemeup log ") {
		return nil, errors.New("not a log command")
	}
	
	// Extract content after "@wrikemeup log "
	content := strings.TrimPrefix(comment, "@wrikemeup log ")
	content = strings.TrimSpace(content)
	
	if content == "" {
		return nil, errors.New("log command requires entries")
	}
	
	// Split by comma, handling quoted descriptions
	rawEntries := splitByComma(content)
	
	var entries []SpecEntry
	for _, raw := range rawEntries {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		
		entry, err := parseSpecEntry(raw)
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

// parseSpecEntry parses a single entry: <date> <duration> ["description"]
func parseSpecEntry(entry string) (SpecEntry, error) {
	entry = strings.TrimSpace(entry)
	
	// Extract description if present (quoted)
	var description string
	quoteStart := strings.Index(entry, `"`)
	if quoteStart >= 0 {
		quoteEnd := strings.LastIndex(entry, `"`)
		if quoteEnd > quoteStart {
			description = entry[quoteStart+1 : quoteEnd]
			entry = strings.TrimSpace(entry[:quoteStart])
		}
	}
	
	// Now parse: <date> <duration>
	// Find the duration (contains h or m)
	parts := strings.Fields(entry)
	if len(parts) < 2 {
		return SpecEntry{}, errors.New("entry must have at least date and duration")
	}
	
	// Find which part is the duration
	durationIdx := -1
	for i, part := range parts {
		if strings.Contains(part, "h") || strings.Contains(part, "m") {
			durationIdx = i
			break
		}
	}
	
	if durationIdx == -1 {
		return SpecEntry{}, errors.New("no duration found (must contain 'h' or 'm')")
	}
	
	// Everything before duration is the date
	dateStr := strings.Join(parts[0:durationIdx], " ")
	durationStr := parts[durationIdx]
	
	// Parse date
	date, err := ParseSpecDate(dateStr)
	if err != nil {
		return SpecEntry{}, fmt.Errorf("invalid date '%s': %w", dateStr, err)
	}
	
	// Parse duration
	hours, err := ParseSpecDuration(durationStr)
	if err != nil {
		return SpecEntry{}, fmt.Errorf("invalid duration '%s': %w", durationStr, err)
	}
	
	// Note: 0h is valid per spec - means delete this entry
	return SpecEntry{
		Date:        date,
		Hours:       hours,
		Description: description,
	}, nil
}

// ParseSpecDate parses date according to specification.
func ParseSpecDate(dateStr string) (string, error) {
	now := time.Now()
	dateStr = strings.ToLower(strings.TrimSpace(dateStr))
	
	// today
	if dateStr == "today" {
		return now.Format("2006-01-02"), nil
	}
	
	// yesterday
	if dateStr == "yesterday" {
		return now.AddDate(0, 0, -1).Format("2006-01-02"), nil
	}
	
	// last <weekday>
	if strings.HasPrefix(dateStr, "last ") {
		weekdayStr := strings.TrimPrefix(dateStr, "last ")
		return parseLastWeekday(weekdayStr, now)
	}
	
	// weekday names
	if weekday := parseWeekday(dateStr); weekday != -1 {
		return findRecentWeekday(weekday, now), nil
	}
	
	// ISO format: 2024-02-15
	if matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, dateStr); matched {
		return dateStr, nil
	}
	
	// Text month: feb 15, 15 feb, march 20, 20 march
	if date := parseTextMonth(dateStr, now); date != "" {
		return date, nil
	}
	
	// Day only: 15
	if matched, _ := regexp.MatchString(`^\d{1,2}$`, dateStr); matched {
		day, _ := strconv.Atoi(dateStr)
		return fmt.Sprintf("%d-%02d-%02d", now.Year(), now.Month(), day), nil
	}
	
	// Month-day: 02-15
	if matched, _ := regexp.MatchString(`^\d{2}-\d{2}$`, dateStr); matched {
		return fmt.Sprintf("%d-%s", now.Year(), dateStr), nil
	}
	
	return "", fmt.Errorf("unrecognized date format: %s", dateStr)
}

// ParseSpecDuration parses duration according to specification.
// Supports: 3h, 30m, 2h30m
func ParseSpecDuration(durationStr string) (float64, error) {
	durationStr = strings.ToLower(strings.TrimSpace(durationStr))
	
	// Combined: 2h30m
	combinedRegex := regexp.MustCompile(`^(\d+)h(\d+)m$`)
	if matches := combinedRegex.FindStringSubmatch(durationStr); len(matches) == 3 {
		hours, _ := strconv.ParseFloat(matches[1], 64)
		minutes, _ := strconv.ParseFloat(matches[2], 64)
		return hours + (minutes / 60.0), nil
	}
	
	// Hours: 3h, 4.5h
	if strings.HasSuffix(durationStr, "h") {
		hoursStr := strings.TrimSuffix(durationStr, "h")
		hours, err := strconv.ParseFloat(hoursStr, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid hours: %s", hoursStr)
		}
		return hours, nil
	}
	
	// Minutes: 30m
	if strings.HasSuffix(durationStr, "m") {
		minutesStr := strings.TrimSuffix(durationStr, "m")
		minutes, err := strconv.ParseFloat(minutesStr, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid minutes: %s", minutesStr)
		}
		return minutes / 60.0, nil
	}
	
	return 0, fmt.Errorf("invalid duration format: %s", durationStr)
}

// Helper functions

func splitByComma(content string) []string {
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
	
	if current.Len() > 0 {
		entries = append(entries, current.String())
	}
	
	return entries
}

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

func findRecentWeekday(weekday time.Weekday, now time.Time) string {
	currentWeekday := now.Weekday()
	daysBack := int(currentWeekday - weekday)
	
	if daysBack < 0 {
		daysBack += 7
	}
	
	if daysBack == 0 {
		return now.Format("2006-01-02")
	}
	
	targetDate := now.AddDate(0, 0, -daysBack)
	return targetDate.Format("2006-01-02")
}

func parseLastWeekday(weekdayStr string, now time.Time) (string, error) {
	weekday := parseWeekday(weekdayStr)
	if weekday == -1 {
		return "", fmt.Errorf("invalid weekday: %s", weekdayStr)
	}
	
	currentWeekday := now.Weekday()
	daysBack := int(currentWeekday - weekday)
	
	if daysBack <= 0 {
		daysBack += 7
	}
	
	targetDate := now.AddDate(0, 0, -daysBack)
	return targetDate.Format("2006-01-02"), nil
}

func parseTextMonth(dateStr string, now time.Time) string {
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
	
	// Try "month day": feb 15
	if month, ok := months[parts[0]]; ok {
		if day, err := strconv.Atoi(parts[1]); err == nil && day >= 1 && day <= 31 {
			return fmt.Sprintf("%d-%02d-%02d", now.Year(), month, day)
		}
	}
	
	// Try "day month": 15 feb
	if month, ok := months[parts[1]]; ok {
		if day, err := strconv.Atoi(parts[0]); err == nil && day >= 1 && day <= 31 {
			return fmt.Sprintf("%d-%02d-%02d", now.Year(), month, day)
		}
	}
	
	return ""
}

// ParseSpecDelete parses delete command.
// Format: @wrikemeup delete <dates>
func ParseSpecDelete(comment string) ([]string, error) {
	comment = strings.TrimSpace(comment)
	
	if !strings.HasPrefix(comment, "@wrikemeup delete ") {
		return nil, errors.New("not a delete command")
	}
	
	content := strings.TrimPrefix(comment, "@wrikemeup delete ")
	content = strings.TrimSpace(content)
	
	if content == "" {
		return nil, errors.New("delete command requires dates")
	}
	
	// Split by comma
	dateSpecs := strings.Split(content, ",")
	var dates []string
	
	for _, spec := range dateSpecs {
		spec = strings.TrimSpace(spec)
		if spec == "" {
			continue
		}
		
		date, err := ParseSpecDate(spec)
		if err != nil {
			return nil, fmt.Errorf("invalid date '%s': %w", spec, err)
		}
		
		dates = append(dates, date)
	}
	
	if len(dates) == 0 {
		return nil, errors.New("no valid dates found")
	}
	
	return dates, nil
}
