package env

import (
	"errors"
	"os"

	"github.com/lopezator/wrikemeup/internal/wrikemeup"
)

// Retrieve retrieves the configuration from environment variables.
func Retrieve() (*wrikemeup.Config, error) {
	// Get the data from the environment variables.
	usersEnv := os.Getenv("USERS")
	if usersEnv == "" {
		return nil, errors.New("env: missing USERS environment variable")
	}
	githubUsername := os.Getenv("GITHUB_USERNAME")
	if githubUsername == "" {
		return nil, errors.New("env: missing GITHUB_USERNAME environment variable")
	}
	gitHubCommentBody := os.Getenv("GITHUB_COMMENT_BODY")
	if gitHubCommentBody == "" {
		return nil, errors.New("env: missing GITHUB_COMMENT_BODY environment variable")
	}
	gitHubBotToken := os.Getenv("BOT_TOKEN")
	if gitHubBotToken == "" {
		return nil, errors.New("env: missing BOT_TOKEN environment variable")
	}
	gitHubRepo := os.Getenv("GITHUB_REPO")
	if gitHubRepo == "" {
		return nil, errors.New("env: missing GITHUB_REPO environment variable")
	}
	gitHubIssueNumber := os.Getenv("GITHUB_ISSUE_NUMBER")
	if gitHubIssueNumber == "" {
		return nil, errors.New("env: missing GITHUB_ISSUE_NUMBER environment variable")
	}
	return &wrikemeup.Config{
		Users:             usersEnv,
		GitHubUsername:    githubUsername,
		GitHubCommentBody: gitHubCommentBody,
		GitHubBotToken:    gitHubBotToken,
		GitHubRepo:        gitHubRepo,
		GitHubIssueNumber: gitHubIssueNumber,
	}, nil
}
