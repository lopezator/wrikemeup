package github

import (
	"errors"
	"regexp"
)

// regexp to match the comment format.
var re = regexp.MustCompile(`@wrikemeup log ([A-Za-z0-9_-]+)`)

// ParseComment parses the GitHub comment to extract the task ID.
func ParseComment(comment string) (string, error) {
	matches := re.FindStringSubmatch(comment)
	if len(matches) < 2 {
		return "", errors.New("github: Task ID not found in comment. Make sure it follows '@wrikemeup log <task-id>'")
	}
	return matches[1], nil
}