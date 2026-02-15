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

// TimeLog represents a Wrike time log entry.
type TimeLog struct {
	ID          string  `json:"id"`
	TaskID      string  `json:"taskId"`
	UserID      string  `json:"userId"`
	Hours       float64 `json:"hours"`
	TrackedDate string  `json:"trackedDate"`
	Comment     string  `json:"comment"`
	CreatedDate string  `json:"createdDate"`
}

// TimeLogResponse represents the Wrike API response for time logs.
type TimeLogResponse struct {
	Data []TimeLog `json:"data"`
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

// GetTimeLogsStructured retrieves structured time logs for a task.
func (c *Client) GetTimeLogsStructured(wrikeTaskID string) ([]TimeLog, error) {
	url := fmt.Sprintf("https://app-eu.wrike.com/api/v4/tasks/%s/timelogs", wrikeTaskID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("wrike: failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("wrike: API call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("wrike: API returned error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var response TimeLogResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("wrike: failed to decode response: %w", err)
	}

	return response.Data, nil
}

// UpdateTimeLog updates an existing time log entry.
func (c *Client) UpdateTimeLog(timelogID string, hours float64, comment string) error {
	url := fmt.Sprintf("https://app-eu.wrike.com/api/v4/timelogs/%s", timelogID)

	payload := map[string]interface{}{
		"hours":   hours,
		"comment": comment,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("wrike: failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonPayload))
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

// DeleteTimeLog deletes a time log entry.
func (c *Client) DeleteTimeLog(timelogID string) error {
	url := fmt.Sprintf("https://app-eu.wrike.com/api/v4/timelogs/%s", timelogID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("wrike: failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

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

// SyncDailyHoursWithTracking syncs hours to Wrike with tracking for updates/deletes.
// Returns a summary of changes made.
func (c *Client) SyncDailyHoursWithTracking(wrikeTaskID string, newDailyHours map[string]float64, comment string) (map[string]string, error) {
	changes := make(map[string]string)

	// Get existing time logs from Wrike
	existingLogs, err := c.GetTimeLogsStructured(wrikeTaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing time logs: %w", err)
	}

	// Build map of existing logs by date (only those from this bot)
	existingByDate := make(map[string]TimeLog)
	for _, timeLog := range existingLogs {
		// Only track logs from this bot (check if comment contains our marker)
		if timeLog.Comment != "" {
			existingByDate[timeLog.TrackedDate] = timeLog
		}
	}

	// Process new hours
	for date, hours := range newDailyHours {
		if existing, found := existingByDate[date]; found {
			// Entry exists - check if we need to update
			if existing.Hours != hours {
				// Update existing entry
				if err := c.UpdateTimeLog(existing.ID, hours, fmt.Sprintf("%s (Updated)", comment)); err != nil {
					log.Printf("Warning: failed to update time log for %s: %v", date, err)
				} else {
					changes[date] = fmt.Sprintf("Updated: %.2fh → %.2fh", existing.Hours, hours)
				}
			} else {
				changes[date] = fmt.Sprintf("Unchanged: %.2fh", hours)
			}
			delete(existingByDate, date) // Mark as processed
		} else {
			// New entry
			if err := c.LogHoursForDate(wrikeTaskID, hours, date, comment); err != nil {
				log.Printf("Warning: failed to log hours for %s: %v", date, err)
			} else {
				changes[date] = fmt.Sprintf("Added: %.2fh", hours)
			}
		}
	}

	// Any remaining entries in existingByDate should be deleted
	for date, timeLog := range existingByDate {
		if err := c.DeleteTimeLog(timeLog.ID); err != nil {
			log.Printf("Warning: failed to delete time log for %s: %v", date, err)
		} else {
			changes[date] = fmt.Sprintf("Deleted: %.2fh", timeLog.Hours)
		}
	}

	return changes, nil
}

// CompleteTask marks a Wrike task as complete.
// Per specification: Set task status to complete/closed, keep task (don't delete), preserve all logged hours.
func (c *Client) CompleteTask(taskID string) error {
	url := fmt.Sprintf("https://app-eu.wrike.com/api/v4/tasks/%s", taskID)
	
	// Wrike API: Set status to "Completed"
	payload := map[string]interface{}{
		"status": "Completed",
	}
	
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("wrike: failed to marshal payload: %w", err)
	}
	
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonPayload))
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
	
	log.Printf("Successfully marked Wrike task %s as complete", taskID)
	return nil
}
