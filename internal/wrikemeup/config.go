package wrikemeup

// Config holds the configuration for the application.
type Config struct {
	Users               string
	GitHubUsername      string
	GitHubCommentBody   string // Optional, only for bot commands
	GitHubBotToken      string
	GitHubRepo          string
	GitHubIssueNumber   string
	GitHubActionType    string // "auto-link", "sync-hours", "sync-project", or "bot-command"
	WrikeFolderID       string // Optional, for auto-creating tasks
	GitHubProjectID     string // Optional, for project-based workflows
	GitHubProjectItemID string // Optional, for project item updates
	GitHubProjectNumber int    // Optional, project number for queries
}
