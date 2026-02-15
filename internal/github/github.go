package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Client represents a GitHub client that can be used to interact with the GitHub API.
type Client struct {
	*http.Client
	repo     string
	botToken string
}

// NewClient creates a new Wrike client with the provided token.
func NewClient(botToken string, repo string) *Client {
	return &Client{
		Client:   &http.Client{},
		botToken: botToken,
		repo:     repo,
	}
}

// PostComment posts a comment on a GitHub issue using the GitHub bot.
func (c *Client) PostComment(issueNumber string) error {
	comment := "Hey! I just logged my hours on this task. Please check it out."
	return c.PostCommentWithBody(issueNumber, comment)
}

// PostCommentWithBody posts a custom comment on a GitHub issue.
func (c *Client) PostCommentWithBody(issueNumber string, comment string) error {
	commentURL := fmt.Sprintf("https://api.github.com/repos/%s/issues/%s/comments", c.repo, issueNumber)
	commentPayload := map[string]string{
		"body": comment,
	}
	commentBody, err := json.Marshal(commentPayload)
	if err != nil {
		return fmt.Errorf("wrikemeup: error when marshaling JSON contents: %w", err)
	}
	commentReq, err := http.NewRequest("POST", commentURL, bytes.NewBuffer(commentBody))
	if err != nil {
		return fmt.Errorf("error when creating the comment request: %w", err)
	}
	commentReq.Header.Set("Authorization", "Bearer "+c.botToken)
	commentReq.Header.Set("Accept", "application/vnd.github+json")
	commentReq.Header.Set("Content-Type", "application/json")
	commentResp, err := c.Do(commentReq)
	if err != nil {
		return fmt.Errorf("error when making the comment request: %w", err)
	}
	defer func() {
		if err := commentResp.Body.Close(); err != nil {
			log.Printf("wrikemeup: error closing response body: %v", err)
		}
	}()
	if commentResp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(commentResp.Body)
		return fmt.Errorf("wrikemeup: GitHub API error: %s", string(bodyBytes))
	}
	return nil
}
