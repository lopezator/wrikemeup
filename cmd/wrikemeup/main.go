package main

import (
	"fmt"
	"log"

	"github.com/lopezator/wrikemeup/internal/env"
	"github.com/lopezator/wrikemeup/internal/github"
	userpkg "github.com/lopezator/wrikemeup/internal/user"
	"github.com/lopezator/wrikemeup/internal/wrike"
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

	// Build the wrike client.
	wrikeClient := wrike.NewClient(user.WrikeToken)

	// Retrieve the taskID from the GitHub comment.
	wrikeTaskID, err := github.ParseComment(config.GitHubCommentBody)
	if err != nil {
		log.Fatalf("wrikemeup: error retrieving the taskID from the github comment: %v", err)
	}

	// Call the Wrike API to get the time logs for the retrieved task ID.
	timeLogs, err := wrikeClient.GetTimeLogs(wrikeTaskID)
	if err != nil {
		log.Fatalf("wrikemeup: wrike API call failed: %v", err)
	}

	// Build the GitHub client.
	githubClient := github.NewClient(config.GitHubBotToken, config.GitHubRepo)

	// Post a comment on the GitHub issue.
	githubClient.PostComment(config.GitHubIssueNumber)

	// Print the time logs response body.
	fmt.Println(string(timeLogs))
}
