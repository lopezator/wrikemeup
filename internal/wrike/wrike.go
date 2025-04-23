package wrike

import (
	"io"
	"log"
	"net/http"
)

// Client represents a Wrike client that can be used to interact with the Wrike API.
type Client struct {
	*http.Client
	token string
}

// NewClient creates a new Wrike client with the provided token.
func NewClient(wrikeToken string) *Client {
	return &Client{
		Client: &http.Client{},
		token:  wrikeToken,
	}
}

// GetTimeLogs retrieves the time logs for a given Wrike task ID.
func (c *Client) GetTimeLogs(wrikeTaskID string) ([]byte, error) {
	req, err := http.NewRequest("GET", "https://app-eu.wrike.com/api/v4/tasks/"+wrikeTaskID+"/timelogs", nil)
	if err != nil {
		log.Fatal("wrikemeup: failed to create request:", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.Do(req)
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
	return body, nil
}
