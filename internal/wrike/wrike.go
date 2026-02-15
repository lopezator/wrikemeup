package wrike

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Client represents a Wrike client that can be used to interact with the Wrike API.
type Client struct {
	*http.Client
	token string
}

// Task represents a Wrike task response.
type Task struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// TaskResponse represents the Wrike API response for task operations.
type TaskResponse struct {
	Data []Task `json:"data"`
}

// NewClient creates a new Wrike client with the provided token.
func NewClient(wrikeToken string) *Client {
	return &Client{
		Client: &http.Client{},
		token:  wrikeToken,
	}
}

// CreateTask creates a new Wrike task in the specified folder.
func (c *Client) CreateTask(folderID string, title string, description string) (*Task, error) {
	url := fmt.Sprintf("https://app-eu.wrike.com/api/v4/folders/%s/tasks", folderID)

	payload := map[string]interface{}{
		"title":       title,
		"description": description,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("wrike: failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("wrike: failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("wrike: API call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("wrike: API returned error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var taskResp TaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
		return nil, fmt.Errorf("wrike: failed to decode response: %w", err)
	}

	if len(taskResp.Data) == 0 {
		return nil, fmt.Errorf("wrike: no task returned in response")
	}

	return &taskResp.Data[0], nil
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

// LogHours logs hours to a Wrike task for today's date.
func (c *Client) LogHours(wrikeTaskID string, hours float64, comment string) error {
	return c.LogHoursForDate(wrikeTaskID, hours, time.Now().Format("2006-01-02"), comment)
}

// LogHoursForDate logs hours to a Wrike task for a specific date.
func (c *Client) LogHoursForDate(wrikeTaskID string, hours float64, date string, comment string) error {
	url := fmt.Sprintf("https://app-eu.wrike.com/api/v4/tasks/%s/timelogs", wrikeTaskID)

	payload := map[string]interface{}{
		"hours":       hours,
		"trackedDate": date,
		"comment":     comment,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("wrike: failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("wrike: failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("wrike: API call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("wrike: API returned error: %s (status: %d)", string(body), resp.StatusCode)
	}

	return nil
}

// LogDailyHours logs hours across multiple dates to a Wrike task.
func (c *Client) LogDailyHours(wrikeTaskID string, dailyHours map[string]float64, comment string) error {
	for date, hours := range dailyHours {
		dateComment := fmt.Sprintf("%s (Date: %s)", comment, date)
		if err := c.LogHoursForDate(wrikeTaskID, hours, date, dateComment); err != nil {
			return fmt.Errorf("failed to log %v hours for %s: %w", hours, date, err)
		}
	}
	return nil
}
