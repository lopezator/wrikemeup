package env

import (
	"errors"
	"os"
	"strconv"

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
	gitHubBotToken := os.Getenv("BOT_TOKEN")
	if gitHubBotToken == "" {
		return nil, errors.New("env: missing BOT_TOKEN environment variable")
	}
	gitHubRepo := os.Getenv("GITHUB_REPO")
	if gitHubRepo == "" {
		return nil, errors.New("env: missing GITHUB_REPO environment variable")
	}

	// Optional environment variables
	gitHubIssueNumber := os.Getenv("GITHUB_ISSUE_NUMBER")
	gitHubCommentBody := os.Getenv("GITHUB_COMMENT_BODY")
	gitHubActionType := os.Getenv("GITHUB_ACTION_TYPE")
	gitHubIssueAction := os.Getenv("GITHUB_ISSUE_ACTION")
	wrikeFolderID := os.Getenv("WRIKE_FOLDER_ID")
	gitHubProjectID := os.Getenv("GITHUB_PROJECT_ID")
	gitHubProjectItemID := os.Getenv("GITHUB_PROJECT_ITEM_ID")
	gitHubProjectNumberStr := os.Getenv("GITHUB_PROJECT_NUMBER")

	// Parse project number
	var gitHubProjectNumber int
	if gitHubProjectNumberStr != "" {
		var err error
		gitHubProjectNumber, err = strconv.Atoi(gitHubProjectNumberStr)
		if err != nil {
			return nil, errors.New("env: GITHUB_PROJECT_NUMBER must be a number")
		}
	}

	// Default to bot-command if not specified
	if gitHubActionType == "" {
		gitHubActionType = "bot-command"
	}

	return &wrikemeup.Config{
		Users:               usersEnv,
		GitHubUsername:      githubUsername,
		GitHubCommentBody:   gitHubCommentBody,
		GitHubBotToken:      gitHubBotToken,
		GitHubRepo:          gitHubRepo,
		GitHubIssueNumber:   gitHubIssueNumber,
		GitHubActionType:    gitHubActionType,
		GitHubIssueAction:   gitHubIssueAction,
		WrikeFolderID:       wrikeFolderID,
		GitHubProjectID:     gitHubProjectID,
		GitHubProjectItemID: gitHubProjectItemID,
		GitHubProjectNumber: gitHubProjectNumber,
	}, nil
}
