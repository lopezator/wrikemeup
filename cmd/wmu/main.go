package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
)

// User is a struct holding the needed data to sign up any user to the WrikeMeUp service.
type User struct {
	GitHubUsername string `json:"github_username"`
	WrikeEmail     string `json:"wrike_email"`
	WrikeToken     string `json:"wrike_token"`
}

func main() {
	// Get the users list.
	usersEnv := os.Getenv("USERS")
	if usersEnv == "" {
		log.Fatal("wrikemeup: missing USERS environment variable")
	}
	usersJson, err := base64.StdEncoding.DecodeString(usersEnv)
	if err != nil {
		log.Fatalf("wrikemeup: failed to decode USERS: %v", err)
	}
	var users []User
	if err := json.Unmarshal(usersJson, &users); err != nil {
		log.Fatalf("wrikemeup: failed to parse USERS: %v", err)
	}

	// Get the GitHub username and match against the users list.
	githubUsername := os.Getenv("GITHUB_USERNAME")
	if githubUsername == "" {
		log.Fatal("wrikemeup: missing GITHUB_USERNAME environment variable")
	}
	var user *User
	for _, u := range users {
		if u.GitHubUsername == githubUsername {
			user = &u
		}
	}
	if user == nil {
		log.Fatalf("wrikemeup: no credentials found for GitHub user: %s", githubUsername)
	}

	// Get the GitHub comment body and look for the task ID.
	gitHubCommentBody := os.Getenv("GITHUB_COMMENT_BODY")
	if gitHubCommentBody == "" {
		log.Fatal("wrikemeup: missing GITHUB_COMMENT_BODY environment variable")
	}
	re := regexp.MustCompile(`@wrikemeup log ([A-Za-z0-9_-]+)`)
	matches := re.FindStringSubmatch(gitHubCommentBody)
	if len(matches) < 2 {
		log.Fatal("wrikemeup: Task ID not found in comment. Make sure it follows '@wrikemeup log <task-id>'.")
	}
	wrikeTaskID := matches[1]

	// Get the GitHub bot token from the environment variable.
	gitHubBotToken := os.Getenv("BOT_TOKEN")
	if gitHubBotToken == "" {
		log.Fatal("wrikemeup: missing BOT_TOKEN environment variable")
	}

	// Get the GitHub repo from the environment variable.
	gitHubRepo := os.Getenv("GITHUB_REPO")
	if gitHubRepo == "" {
		log.Fatal("wrikemeup: missing GITHUB_REPO environment variable")
	}

	// Get the GitHub issue number from the environment variable.
	gitHubIssueNumber := os.Getenv("GITHUB_ISSUE_NUMBER")
	if gitHubIssueNumber == "" {
		log.Fatal("wrikemeup: missing GITHUB_ISSUE_NUMBER environment variable")
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
	commentURL := fmt.Sprintf("https://api.github.com/repos/%s/issues/%s/comments", gitHubRepo, gitHubIssueNumber)
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
	commentReq.Header.Set("Authorization", "Bearer "+gitHubBotToken)
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
