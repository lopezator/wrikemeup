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
		handleAutoLink(wrikeClient, githubClient, config.GitHubIssueNumber, config, user)
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

	// Get issue number and project data from the project item
	projectItem, issueNumber, err := githubClient.GetIssueFromProjectItem(config.GitHubProjectItemID, config.GitHubProjectNumber)
	if err != nil {
		log.Fatalf("wrikemeup: failed to get issue from project item: %v", err)
	}

	if issueNumber == 0 {
		log.Println("wrikemeup: no issue associated with this project item")
		return
	}

	log.Printf("Processing project item for issue #%d", issueNumber)

	// Check if this is marked as a Wrike parent via custom field
	if projectItem.IsWrikeParent {
		// Auto-create Wrike task if not already linked
		if projectItem.WrikeTaskID == "" {
			handleAutoLink(wrikeClient, githubClient, strconv.Itoa(issueNumber), config, user)
		}
	}

	// If there are hours or a linked task, sync them
	if projectItem.WrikeTaskID != "" || projectItem.Hours > 0 {
		config.GitHubIssueNumber = strconv.Itoa(issueNumber)
		handleSyncHours(wrikeClient, githubClient, config, user)
	}

	fmt.Printf("Successfully processed project item for issue #%d\n", issueNumber)
}

// handleAutoLink automatically creates a Wrike task and links it to the GitHub issue.
func handleAutoLink(wrikeClient *wrike.Client, githubClient *github.Client, issueNumber string, config *wrikemeup.Config, user *userpkg.User) {
	// Get issue metadata
	metadata, err := githubClient.GetIssueMetadata(issueNumber)
	if err != nil {
		log.Fatalf("wrikemeup: failed to get issue metadata: %v", err)
	}

	// Check if already linked
	if metadata.WrikeTaskID != "" {
		log.Printf("Issue #%s is already linked to Wrike task %s", issueNumber, metadata.WrikeTaskID)
		return
	}

	// Check if folder ID is configured
	if config.WrikeFolderID == "" {
		log.Printf("WRIKE_FOLDER_ID not configured, skipping auto-link for issue #%s", issueNumber)
		comment := "⚠️ Cannot auto-create Wrike task: WRIKE_FOLDER_ID not configured. Please use `@wrikemeup link <task-id>` to manually link a task."
		if err := githubClient.PostCommentWithBody(issueNumber, comment); err != nil {
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
	totalDailyHours := make(map[string]float64)
	
	// Copy daily hours if present
	for date, hours := range metadata.DailyHours {
		totalDailyHours[date] = hours
	}

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
			// Aggregate daily hours from children
			for date, hours := range childMetadata.DailyHours {
				totalDailyHours[date] += hours
			}
		}
	}

	// Calculate incremental hours (delta since last sync)
	hoursToLog := totalHours - metadata.LastSyncedHours
	
	// If using daily breakdown, log to specific dates
	if len(totalDailyHours) > 0 {
		// For daily hours, we need to track which dates are new
		// For simplicity, log all daily hours (Wrike handles duplicates)
		comment := fmt.Sprintf("Auto-synced from GitHub issue #%s (aggregated %d child issues)", config.GitHubIssueNumber, len(childIssues))
		if len(childIssues) == 0 {
			comment = fmt.Sprintf("Auto-synced from GitHub issue #%s", config.GitHubIssueNumber)
		}
		
		if err := wrikeClient.LogDailyHours(metadata.WrikeTaskID, totalDailyHours, comment); err != nil {
			log.Fatalf("wrikemeup: failed to log daily hours to Wrike: %v", err)
		}
		
		log.Printf("Successfully synced %d days of hours to Wrike", len(totalDailyHours))
	} else if hoursToLog > 0 {
		// Incremental logging: only log the difference since last sync
		comment := fmt.Sprintf("Auto-synced %.2fh from GitHub issue #%s (aggregated %d child issues)", hoursToLog, config.GitHubIssueNumber, len(childIssues))
		if len(childIssues) == 0 {
			comment = fmt.Sprintf("Auto-synced %.2fh from GitHub issue #%s", hoursToLog, config.GitHubIssueNumber)
		}
		
		if err := wrikeClient.LogHours(metadata.WrikeTaskID, hoursToLog, comment); err != nil {
			log.Fatalf("wrikemeup: failed to log hours to Wrike: %v", err)
		}
		
		// Update the last synced hours in the issue body
		if err := githubClient.UpdateLastSyncedHours(config.GitHubIssueNumber, totalHours); err != nil {
			log.Printf("Warning: failed to update last synced hours: %v", err)
		}
		
		log.Printf("Successfully synced %.2f incremental hours to Wrike", hoursToLog)
	} else {
		log.Printf("No new hours to sync for issue #%s (current: %.2f, last synced: %.2f)", 
			config.GitHubIssueNumber, totalHours, metadata.LastSyncedHours)
		return
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
	case "sync":
		handleSyncCommand(wrikeClient, githubClient, config, user)
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

// handleSyncCommand handles the 'sync' command to manually sync hours to Wrike without closing the issue.
func handleSyncCommand(wrikeClient *wrike.Client, githubClient *github.Client, config *wrikemeup.Config, user *userpkg.User) {
	// This is essentially the same as handleSyncHours, but triggered manually via command
	log.Printf("Manual sync requested for issue #%s", config.GitHubIssueNumber)
	handleSyncHours(wrikeClient, githubClient, config, user)
}
