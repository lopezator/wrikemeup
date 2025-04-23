package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/lopezator/wrikemeup/internal/env"
	"github.com/lopezator/wrikemeup/internal/github"
	userpkg "github.com/lopezator/wrikemeup/internal/user"
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

	// Retrieve the taskID from the github comment.
	wrikeTaskID, err := github.ParseComment(config.GitHubCommentBody)
	if err != nil {
		log.Fatalf("wrikemeup: error retrieving the taskID from the github comment: %v", err)
	}

	// Call the Wrike API to get the time logs for the retrieved task ID.
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://app-eu.wrike.com/api/v4/tasks/%s/timelogs", wrikeTaskID), nil)
	if err != nil {
		log.Fatal("wrikemeup: failed to create request:", err)
	}
	req.Header.Set("Authorization", "Bearer "+user.WrikeToken)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("wrikemeup: wrike API call failed:", err)
	}
	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			log.Fatal("wrikemeup: error closing response body:", err)
		}
	}(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body) // Read the response body for error details
		log.Fatalf("wrikemeup: wrike API returned an error: %s (status code: %d)", string(body), resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("wrikemeup: error reading response body:", err)
	}

	// Write a comment in GitHub with the GitHub bot.
	comment := "Hey! I just logged my hours on this task. Please check it out."
	commentURL := fmt.Sprintf("https://api.github.com/repos/%s/issues/%s/comments", config.GitHubRepo, config.GitHubIssueNumber)
	commentPayload := map[string]string{
		"body": comment,
	}
	commentBody, err := json.Marshal(commentPayload)
	if err != nil {
		log.Fatalf("wrikemeup: error when marshaling JSON contents: %v", err)
	}
	commentReq, err := http.NewRequest("POST", commentURL, bytes.NewBuffer(commentBody))
	if err != nil {
		log.Fatalf("error when creating the comment request: %v", err)
	}
	commentReq.Header.Set("Authorization", "Bearer "+config.GitHubBotToken)
	commentReq.Header.Set("Accept", "application/vnd.github+json")
	commentReq.Header.Set("Content-Type", "application/json")
	commentResp, err := client.Do(commentReq)
	if err != nil {
		log.Fatalf("error when making the comment request: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal("wrikemeup: error closing response body:", err)
		}
	}(commentResp.Body)
	if commentResp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(commentResp.Body)
		log.Fatalf("wrikemeup: GitHub API error: %s", string(bodyBytes))
	}

	// Print the response body
	fmt.Println(string(body))
}
