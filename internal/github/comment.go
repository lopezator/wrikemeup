package github

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

// Command represents a parsed command from a GitHub comment.
type Command struct {
	Action      string  // "log", "link", "loghours"
	TaskID      string  // Wrike task ID
	Hours       float64 // Hours to log
	IssueNumber string  // GitHub issue number for linking
}

// regexp patterns for different command formats.
var (
	reLog      = regexp.MustCompile(`@wrikemeup log ([A-Za-z0-9_-]+)`)
	reLink     = regexp.MustCompile(`@wrikemeup link ([A-Za-z0-9_-]+)`)
	reLogHours = regexp.MustCompile(`@wrikemeup loghours ([A-Za-z0-9_-]+)\s+([\d.]+)h?`)
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

// ParseCommand parses the GitHub comment and returns a Command struct.
func ParseCommand(comment string) (*Command, error) {
	comment = strings.TrimSpace(comment)

	// Try to match loghours command
	if matches := reLogHours.FindStringSubmatch(comment); len(matches) >= 3 {
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

	// Try to match log command
	if matches := reLog.FindStringSubmatch(comment); len(matches) >= 2 {
		return &Command{
			Action: "log",
			TaskID: matches[1],
		}, nil
	}

	return nil, errors.New("github: command not recognized. Use '@wrikemeup log <task-id>', '@wrikemeup link <task-id>', or '@wrikemeup loghours <task-id> <hours>h'")
}