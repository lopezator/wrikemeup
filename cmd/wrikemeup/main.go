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
// Per spec: Also handles close behavior (mark Wrike task complete).
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

	// Handle close behavior per specification
	if config.GitHubIssueAction == "closed" {
		handleCloseIssue(wrikeClient, githubClient, metadata, config)
		return
	}

	// Check for validation errors and post them
	if len(metadata.ValidationErrors) > 0 {
		log.Printf("Found %d validation errors in hours format", len(metadata.ValidationErrors))
		if err := githubClient.PostValidationErrors(config.GitHubIssueNumber, metadata.ValidationErrors); err != nil {
			log.Printf("Warning: failed to post validation errors: %v", err)
		}
		// Still continue with valid hours if any were found
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

	// If using daily breakdown, use smart sync with tracking
	if len(totalDailyHours) > 0 {
		comment := fmt.Sprintf("Auto-synced from GitHub issue #%s", config.GitHubIssueNumber)
		if len(childIssues) > 0 {
			comment = fmt.Sprintf("Auto-synced from GitHub issue #%s (aggregated %d child issues)", config.GitHubIssueNumber, len(childIssues))
		}

		// Use new sync with tracking (handles add/update/delete)
		changes, err := wrikeClient.SyncDailyHoursWithTracking(metadata.WrikeTaskID, totalDailyHours, comment)
		if err != nil {
			log.Fatalf("wrikemeup: failed to sync hours to Wrike: %v", err)
		}

		log.Printf("Successfully synced %d days of hours to Wrike", len(totalDailyHours))

		// Post summary table to GitHub
		if err := githubClient.PostHoursSummary(config.GitHubIssueNumber, totalDailyHours, changes); err != nil {
			log.Printf("Warning: failed to post hours summary: %v", err)
		}
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

// handleBotCommand handles bot commands from comments.
func handleBotCommand(wrikeClient *wrike.Client, githubClient *github.Client, config *wrikemeup.Config, user *userpkg.User) {
	comment := config.GitHubCommentBody
	
	// Parse command
	cmd, err := github.ParseCommand(comment)
	if err != nil {
		// Invalid command - post helpful error message
		errorMsg := "❌ Invalid command format.\n\n📖 See [Bot Commands Reference](https://github.com/lopezator/wrikemeup/blob/main/BOT_COMMANDS.md)"
		if postErr := githubClient.PostCommentWithBody(config.GitHubIssueNumber, errorMsg); postErr != nil {
			log.Printf("Warning: failed to post error message: %v", postErr)
		}
		log.Fatalf("wrikemeup: error parsing command: %v", err)
	}

	// Handle different command actions
	switch cmd.Action {
	case "log":
		if cmd.DailyHours != nil && len(cmd.DailyHours) > 0 {
			// New log format with daily hours
			handleNewLogCommand(wrikeClient, githubClient, cmd, config, user)
		} else {
			// Legacy log format (retrieve time logs)
			handleLegacyLogCommand(wrikeClient, githubClient, cmd, config)
		}
	case "link":
		handleLinkCommand(githubClient, cmd, config)
	case "loghours":
		handleLogHoursCommand(wrikeClient, githubClient, cmd, config, user)
	case "sync":
		handleSyncCommand(wrikeClient, githubClient, config, user)
	case "delete":
		handleDeleteCommand(wrikeClient, githubClient, cmd, config, user)
	case "show":
		handleShowCommand(githubClient, config)
	default:
		log.Fatalf("wrikemeup: unknown command action: %s", cmd.Action)
	}
}

// handleNewLogCommand handles the new 'log' command with relative dates.
func handleNewLogCommand(wrikeClient *wrike.Client, githubClient *github.Client, cmd *github.Command, config *wrikemeup.Config, user *userpkg.User) {
	// Get issue metadata
	metadata, err := githubClient.GetIssueMetadata(config.GitHubIssueNumber)
	if err != nil {
		log.Fatalf("wrikemeup: failed to get issue metadata: %v", err)
	}

	// Check if linked to Wrike task
	if metadata.WrikeTaskID == "" {
		errorMsg := "❌ This issue is not linked to a Wrike task.\n\nPlease add the `wrike-parent` label or use `@wrikemeup link <task-id>` first."
		if err := githubClient.PostCommentWithBody(config.GitHubIssueNumber, errorMsg); err != nil {
			log.Printf("Warning: failed to post error: %v", err)
		}
		log.Fatal("wrikemeup: no Wrike task linked to this issue")
	}

	// Sync the new hours to Wrike
	comment := fmt.Sprintf("Logged from @wrikemeup command")
	changes, err := wrikeClient.SyncDailyHoursWithTracking(metadata.WrikeTaskID, cmd.DailyHours, comment)
	if err != nil {
		errorMsg := fmt.Sprintf("❌ Failed to log hours to Wrike: %v", err)
		if postErr := githubClient.PostCommentWithBody(config.GitHubIssueNumber, errorMsg); postErr != nil {
			log.Printf("Warning: failed to post error: %v", postErr)
		}
		log.Fatalf("wrikemeup: failed to log hours: %v", err)
	}

	log.Printf("Successfully logged %d day(s) of hours to Wrike", len(cmd.DailyHours))

	// Get ALL hours from comments to show complete state
	allHours, err := githubClient.GetAllLoggedHours(config.GitHubIssueNumber)
	if err != nil {
		log.Printf("Warning: failed to get all hours: %v", err)
		allHours = cmd.DailyHours // Fallback to just the new hours
	}

	// Post summary table
	if err := githubClient.PostHoursSummary(config.GitHubIssueNumber, allHours, changes); err != nil {
		log.Printf("Warning: failed to post hours summary: %v", err)
	}

	fmt.Printf("Successfully logged hours to Wrike task %s\n", metadata.WrikeTaskID)
}

// handleDeleteCommand handles the 'delete' command to remove hours for specific dates.
func handleDeleteCommand(wrikeClient *wrike.Client, githubClient *github.Client, cmd *github.Command, config *wrikemeup.Config, user *userpkg.User) {
	// Get issue metadata
	metadata, err := githubClient.GetIssueMetadata(config.GitHubIssueNumber)
	if err != nil {
		log.Fatalf("wrikemeup: failed to get issue metadata: %v", err)
	}

	// Check if linked to Wrike task
	if metadata.WrikeTaskID == "" {
		errorMsg := "❌ This issue is not linked to a Wrike task."
		if err := githubClient.PostCommentWithBody(config.GitHubIssueNumber, errorMsg); err != nil {
			log.Printf("Warning: failed to post error: %v", err)
		}
		log.Fatal("wrikemeup: no Wrike task linked to this issue")
	}

	// Get current time logs from Wrike
	timeLogs, err := wrikeClient.GetTimeLogsStructured(metadata.WrikeTaskID)
	if err != nil {
		errorMsg := fmt.Sprintf("❌ Failed to get time logs from Wrike: %v", err)
		if postErr := githubClient.PostCommentWithBody(config.GitHubIssueNumber, errorMsg); postErr != nil {
			log.Printf("Warning: failed to post error: %v", postErr)
		}
		log.Fatalf("wrikemeup: failed to get time logs: %v", err)
	}

	// Delete entries for specified dates
	changes := make(map[string]string)
	for _, date := range cmd.DeleteDates {
		// Find time log for this date
		found := false
		for _, timeLog := range timeLogs {
			if timeLog.TrackedDate == date {
				// Delete the time log
				if err := wrikeClient.DeleteTimeLog(timeLog.ID); err != nil {
					log.Printf("Warning: failed to delete time log for %s: %v", date, err)
					changes[date] = fmt.Sprintf("Delete failed: %v", err)
				} else {
					changes[date] = fmt.Sprintf("Deleted: %.2fh", timeLog.Hours)
					found = true
				}
				break
			}
		}
		if !found {
			changes[date] = "Not found (already deleted?)"
		}
	}

	log.Printf("Deleted hours for %d date(s)", len(cmd.DeleteDates))

	// Get remaining hours to show current state
	remainingHours, err := githubClient.GetAllLoggedHours(config.GitHubIssueNumber)
	if err != nil {
		log.Printf("Warning: failed to get remaining hours: %v", err)
		remainingHours = make(map[string]float64)
	}

	// Post summary
	if err := githubClient.PostHoursSummary(config.GitHubIssueNumber, remainingHours, changes); err != nil {
		log.Printf("Warning: failed to post summary: %v", err)
	}

	fmt.Printf("Successfully deleted hours for dates: %v\n", cmd.DeleteDates)
}

// handleShowCommand handles the 'show' command to display current logged hours.
func handleShowCommand(githubClient *github.Client, config *wrikemeup.Config) {
	// Get all logged hours from comments
	allHours, err := githubClient.GetAllLoggedHours(config.GitHubIssueNumber)
	if err != nil {
		errorMsg := fmt.Sprintf("❌ Failed to retrieve logged hours: %v", err)
		if postErr := githubClient.PostCommentWithBody(config.GitHubIssueNumber, errorMsg); postErr != nil {
			log.Printf("Warning: failed to post error: %v", postErr)
		}
		log.Fatalf("wrikemeup: failed to get logged hours: %v", err)
	}

	// Post summary with no changes (just current state)
	if err := githubClient.PostHoursSummary(config.GitHubIssueNumber, allHours, make(map[string]string)); err != nil {
		log.Printf("Warning: failed to post summary: %v", err)
	}

	fmt.Printf("Displayed current logged hours for issue #%s\n", config.GitHubIssueNumber)
}

// handleLegacyLogCommand handles the legacy 'log' command to retrieve time logs.
func handleLegacyLogCommand(wrikeClient *wrike.Client, githubClient *github.Client, cmd *github.Command, config *wrikemeup.Config) {
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

// handleSpecLogCommand handles the specification-compliant log command.
// Implements the full scan aggregation algorithm from the specification.
// handleCloseIssue handles closing an issue and marking the Wrike task complete.
func handleCloseIssue(wrikeClient *wrike.Client, githubClient *github.Client, metadata *github.IssueMetadata, config *wrikemeup.Config) {
	log.Printf("Handling close event for issue #%s", config.GitHubIssueNumber)

	// Check if issue has Wrike ID
	if metadata.WrikeTaskID == "" {
		log.Printf("Issue #%s has no Wrike ID, no action needed on close", config.GitHubIssueNumber)
		return
	}

	// Mark Wrike task complete
	log.Printf("Marking Wrike task %s as complete for issue #%s", metadata.WrikeTaskID, config.GitHubIssueNumber)

	if err := wrikeClient.CompleteTask(metadata.WrikeTaskID); err != nil {
		log.Printf("Warning: failed to mark Wrike task as complete: %v", err)
		return
	}

	// Post success comment
	comment := fmt.Sprintf("✅ Marked Wrike task %s as complete.\n\nAll logged hours have been preserved.", metadata.WrikeTaskID)
	if err := githubClient.PostCommentWithBody(config.GitHubIssueNumber, comment); err != nil {
		log.Printf("Warning: failed to post comment: %v", err)
	}

	log.Printf("Successfully marked Wrike task %s as complete", metadata.WrikeTaskID)
	fmt.Printf("Issue #%s closed - Wrike task %s marked complete\n", config.GitHubIssueNumber, metadata.WrikeTaskID)
}
