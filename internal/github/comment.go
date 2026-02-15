package github

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Now returns the current time. Extracted as a variable for testing.
var Now = time.Now

// Command represents a parsed command from a GitHub comment.
type Command struct {
	Action      string       // "log", "delete", "show"
	LogEntries  []SpecEntry  // Parsed log entries from spec format
	DeleteDates []string     // Dates to delete
}

// regexp patterns for commands
var (
	reShow = regexp.MustCompile(`@wrikemeup show`)
)

const (
	helpMessage = `Commands:
- @wrikemeup log today 3h "description"
- @wrikemeup log today 3h, yesterday 2h
- @wrikemeup delete monday
- @wrikemeup show

See: https://github.com/lopezator/wrikemeup`
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

	// Try to match log command (spec format)
	if strings.HasPrefix(comment, "@wrikemeup log ") {
		entries, err := ParseSpecLog(comment)
		if err != nil {
			return nil, fmt.Errorf("invalid log format: %w\n\n%s", err, helpMessage)
		}

		return &Command{
			Action:     "log",
			LogEntries: entries,
		}, nil
	}

	// Try to match delete command (spec format)
	if strings.HasPrefix(comment, "@wrikemeup delete ") {
		dates, err := ParseSpecDelete(comment)
		if err != nil {
			return nil, fmt.Errorf("invalid delete format: %w\n\n%s", err, helpMessage)
		}

		return &Command{
			Action:      "delete",
			DeleteDates: dates,
		}, nil
	}

	return nil, errors.New("command not recognized.\n\n" + helpMessage)
}

