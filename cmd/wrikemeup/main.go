package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/lopezator/wrikemeup/internal/env"
	"github.com/lopezator/wrikemeup/internal/github"
	userpkg "github.com/lopezator/wrikemeup/internal/user"
	"github.com/lopezator/wrikemeup/internal/wrike"
	"github.com/lopezator/wrikemeup/internal/wrikemeup"
)

func main() {
	// Retrieve the configuration from environment variables.
	config, err := env.Retrieve()
	if err != nil {
		log.Fatalf("wrikemeup: error retrieving environment variables: %v", err)
	}

	// Get the user.
	user, err := userpkg.DecodeUserFromEnv(config.GitHubUsername, config.Users)
	if err != nil {
		log.Fatalf("wrikemeup: error decoding the user from the users environment variable: %v", err)
	}

	// Build the GitHub client.
	githubClient := github.NewClient(config.GitHubBotToken, config.GitHubRepo)

	// Build the wrike client.
	wrikeClient := wrike.NewClient(user.WrikeToken)

	// Handle different action types
	switch config.GitHubActionType {
	case "sync-project":
		handleSyncProject(wrikeClient, githubClient, config, user)
	case "auto-link":
		handleAutoLink(wrikeClient, githubClient, config, user)
	case "sync-hours":
		handleSyncHours(wrikeClient, githubClient, config, user)
	case "bot-command":
		handleBotCommand(wrikeClient, githubClient, config, user)
	default:
		log.Fatalf("wrikemeup: unknown action type: %s", config.GitHubActionType)
	}
}

// handleSyncProject handles project item updates (when custom fields change).
func handleSyncProject(wrikeClient *wrike.Client, githubClient *github.Client, config *wrikemeup.Config, user *userpkg.User) {
	if config.GitHubProjectNumber == 0 {
		log.Fatal("wrikemeup: GITHUB_PROJECT_NUMBER not configured")
	}

	// Note: GitHub doesn't directly provide issue number in projects_v2_item event
	// We need to query the project item to get the issue number
	// For now, log a message - in production, you'd query the GraphQL API to get the content_node_id
	log.Printf("Project item %s updated in project %s", config.GitHubProjectItemID, config.GitHubProjectID)
	log.Println("Note: Full project sync implementation requires querying the issue number from the project item")

	// This would require additional GraphQL queries to:
	// 1. Get the issue number from the project item
	// 2. Read the custom field values
	// 3. Process based on field changes
	fmt.Println("Project sync feature ready - requires issue number extraction from project item")
}

// handleAutoLink automatically creates a Wrike task and links it to the GitHub issue.
func handleAutoLink(wrikeClient *wrike.Client, githubClient *github.Client, config *wrikemeup.Config, user *userpkg.User) {
	// Get issue metadata
	metadata, err := githubClient.GetIssueMetadata(config.GitHubIssueNumber)
	if err != nil {
		log.Fatalf("wrikemeup: failed to get issue metadata: %v", err)
	}

	// Check if already linked
	if metadata.WrikeTaskID != "" {
		log.Printf("Issue #%s is already linked to Wrike task %s", config.GitHubIssueNumber, metadata.WrikeTaskID)
		return
	}

	// Check if folder ID is configured
	if config.WrikeFolderID == "" {
		log.Printf("WRIKE_FOLDER_ID not configured, skipping auto-link for issue #%s", config.GitHubIssueNumber)
		comment := "⚠️ Cannot auto-create Wrike task: WRIKE_FOLDER_ID not configured. Please use `@wrikemeup link <task-id>` to manually link a task."
		if err := githubClient.PostCommentWithBody(config.GitHubIssueNumber, comment); err != nil {
			log.Printf("Warning: failed to post comment: %v", err)
		}
		return
	}

	// Create Wrike task
	description := fmt.Sprintf("Auto-created from GitHub issue #%s\n\n%s", config.GitHubIssueNumber, metadata.Body)
	task, err := wrikeClient.CreateTask(config.WrikeFolderID, metadata.Title, description)
	if err != nil {
		log.Fatalf("wrikemeup: failed to create Wrike task: %v", err)
	}

	// Link the task to the issue
	if err := githubClient.AddOrUpdateWrikeTaskID(config.GitHubIssueNumber, task.ID); err != nil {
		log.Fatalf("wrikemeup: failed to link issue to Wrike task: %v", err)
	}

	// Post success comment
	comment := fmt.Sprintf("✅ Automatically created and linked Wrike task: %s\n\nYou can now:\n- Add hours to this issue using `Hours: X.Xh` in the issue body\n- Reference subtasks using `#123`\n- Hours will be synced when the issue is updated or closed", task.ID)
	if err := githubClient.PostCommentWithBody(config.GitHubIssueNumber, comment); err != nil {
		log.Printf("Warning: failed to post comment: %v", err)
	}

	fmt.Printf("Successfully created Wrike task %s and linked to issue #%s\n", task.ID, config.GitHubIssueNumber)
}

// handleSyncHours syncs hours from the GitHub issue to the Wrike task.
func handleSyncHours(wrikeClient *wrike.Client, githubClient *github.Client, config *wrikemeup.Config, user *userpkg.User) {
	issueNum, err := strconv.Atoi(config.GitHubIssueNumber)
	if err != nil {
		log.Fatalf("wrikemeup: invalid issue number: %v", err)
	}

	// Get issue metadata
	metadata, err := githubClient.GetIssueMetadata(config.GitHubIssueNumber)
	if err != nil {
		log.Fatalf("wrikemeup: failed to get issue metadata: %v", err)
	}

	// Check if issue is linked to a Wrike task
	if metadata.WrikeTaskID == "" {
		log.Printf("Issue #%s is not linked to a Wrike task, skipping sync", config.GitHubIssueNumber)
		return
	}

	// Calculate total hours
	totalHours := metadata.Hours

	// Automatically find ALL child issues that reference this parent
	childIssues, err := githubClient.GetChildIssues(issueNum)
	if err != nil {
		log.Printf("Warning: failed to get child issues: %v", err)
		childIssues = []int{}
	}

	// Aggregate hours from all child issues
	if len(childIssues) > 0 {
		log.Printf("Found %d child issues for parent #%d", len(childIssues), issueNum)
		for _, childNum := range childIssues {
			childMetadata, err := githubClient.GetIssueMetadata(strconv.Itoa(childNum))
			if err != nil {
				log.Printf("Warning: failed to get metadata for child issue #%d: %v", childNum, err)
				continue
			}
			if childMetadata.Hours > 0 {
				log.Printf("  - Issue #%d: %.2fh", childNum, childMetadata.Hours)
				totalHours += childMetadata.Hours
			}
		}
	}

	// Only log if there are hours to log
	if totalHours == 0 {
		log.Printf("No hours to sync for issue #%s", config.GitHubIssueNumber)
		return
	}

	// Log hours to Wrike
	comment := fmt.Sprintf("Auto-synced from GitHub issue #%s (aggregated %d child issues)", config.GitHubIssueNumber, len(childIssues))
	if len(childIssues) == 0 {
		comment = fmt.Sprintf("Auto-synced from GitHub issue #%s", config.GitHubIssueNumber)
	}

	if err := wrikeClient.LogHours(metadata.WrikeTaskID, totalHours, comment); err != nil {
		log.Fatalf("wrikemeup: failed to log hours to Wrike: %v", err)
	}

	// Post success comment
	successComment := fmt.Sprintf("✅ Synced %.2fh to Wrike task %s", totalHours, metadata.WrikeTaskID)
	if len(childIssues) > 0 {
		successComment += fmt.Sprintf(" (aggregated from %d child issues)", len(childIssues))
	}
	if err := githubClient.PostCommentWithBody(config.GitHubIssueNumber, successComment); err != nil {
		log.Printf("Warning: failed to post comment: %v", err)
	}

	fmt.Printf("Successfully synced %.2f hours to Wrike task %s\n", totalHours, metadata.WrikeTaskID)
}

// handleBotCommand handles legacy bot commands from comments.
func handleBotCommand(wrikeClient *wrike.Client, githubClient *github.Client, config *wrikemeup.Config, user *userpkg.User) {
	// Parse the command from the GitHub comment.
	cmd, err := github.ParseCommand(config.GitHubCommentBody)
	if err != nil {
		log.Fatalf("wrikemeup: error parsing command: %v", err)
	}

	// Handle different command actions
	switch cmd.Action {
	case "log":
		handleLogCommand(wrikeClient, githubClient, cmd, config)
	case "link":
		handleLinkCommand(githubClient, cmd, config)
	case "loghours":
		handleLogHoursCommand(wrikeClient, githubClient, cmd, config, user)
	default:
		log.Fatalf("wrikemeup: unknown command action: %s", cmd.Action)
	}
}

// handleLogCommand handles the 'log' command to retrieve time logs.
func handleLogCommand(wrikeClient *wrike.Client, githubClient *github.Client, cmd *github.Command, config *wrikemeup.Config) {
	// Call the Wrike API to get the time logs for the retrieved task ID.
	timeLogs, err := wrikeClient.GetTimeLogs(cmd.TaskID)
	if err != nil {
		log.Fatalf("wrikemeup: wrike API call failed: %v", err)
	}

	// Post a comment on the GitHub issue.
	if err := githubClient.PostComment(config.GitHubIssueNumber); err != nil {
		log.Fatalf("wrikemeup: failed to post comment: %v", err)
	}

	// Print the time logs response body.
	fmt.Println(string(timeLogs))
}

// handleLinkCommand handles the 'link' command to link a GitHub issue with a Wrike task.
func handleLinkCommand(githubClient *github.Client, cmd *github.Command, config *wrikemeup.Config) {
	// Add or update the Wrike Task ID in the issue body.
	if err := githubClient.AddOrUpdateWrikeTaskID(config.GitHubIssueNumber, cmd.TaskID); err != nil {
		log.Fatalf("wrikemeup: failed to link issue to Wrike task: %v", err)
	}

	// Post a success comment.
	comment := fmt.Sprintf("✅ Successfully linked this issue to Wrike task: %s", cmd.TaskID)
	if err := githubClient.PostCommentWithBody(config.GitHubIssueNumber, comment); err != nil {
		log.Fatalf("wrikemeup: failed to post comment: %v", err)
	}

	fmt.Printf("Successfully linked issue #%s to Wrike task %s\n", config.GitHubIssueNumber, cmd.TaskID)
}

// handleLogHoursCommand handles the 'loghours' command to log hours to a Wrike task.
func handleLogHoursCommand(wrikeClient *wrike.Client, githubClient *github.Client, cmd *github.Command, config *wrikemeup.Config, user *userpkg.User) {
	// Get issue metadata to check if it has a linked Wrike task
	var taskID string
	if cmd.TaskID != "" {
		taskID = cmd.TaskID
	} else {
		// Try to get the Wrike task ID from the issue metadata
		metadata, err := githubClient.GetIssueMetadata(config.GitHubIssueNumber)
		if err != nil {
			log.Fatalf("wrikemeup: failed to get issue metadata: %v", err)
		}
		if metadata.WrikeTaskID == "" {
			log.Fatalf("wrikemeup: no Wrike task linked to this issue. Use '@wrikemeup link <task-id>' first")
		}
		taskID = metadata.WrikeTaskID
	}

	// Check if this is a parent issue with sub-issues
	metadata, err := githubClient.GetIssueMetadata(config.GitHubIssueNumber)
	if err != nil {
		log.Fatalf("wrikemeup: failed to get issue metadata: %v", err)
	}

	totalHours := cmd.Hours
	comment := fmt.Sprintf("Logged from GitHub issue #%s by %s", config.GitHubIssueNumber, user.GitHubUsername)

	// If there are sub-issues, aggregate their hours
	if len(metadata.SubIssues) > 0 {
		for _, subIssueNum := range metadata.SubIssues {
			subMetadata, err := githubClient.GetIssueMetadata(strconv.Itoa(subIssueNum))
			if err != nil {
				log.Printf("wrikemeup: warning: failed to get metadata for sub-issue #%d: %v", subIssueNum, err)
				continue
			}
			if subMetadata.Hours > 0 {
				totalHours += subMetadata.Hours
			}
		}
		comment = fmt.Sprintf("Logged from GitHub issue #%s (including %d sub-issues) by %s",
			config.GitHubIssueNumber, len(metadata.SubIssues), user.GitHubUsername)
	}

	// Log hours to Wrike
	if err := wrikeClient.LogHours(taskID, totalHours, comment); err != nil {
		log.Fatalf("wrikemeup: failed to log hours to Wrike: %v", err)
	}

	// Post a success comment
	successComment := fmt.Sprintf("✅ Successfully logged %.2f hours to Wrike task %s", totalHours, taskID)
	if len(metadata.SubIssues) > 0 {
		successComment += fmt.Sprintf(" (aggregated from %d sub-issues)", len(metadata.SubIssues))
	}
	if err := githubClient.PostCommentWithBody(config.GitHubIssueNumber, successComment); err != nil {
		log.Fatalf("wrikemeup: failed to post comment: %v", err)
	}

	fmt.Printf("Successfully logged %.2f hours to Wrike task %s\n", totalHours, taskID)
}
