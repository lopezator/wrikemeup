package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// ProjectItem represents a GitHub Projects V2 item.
type ProjectItem struct {
	ID            string
	ProjectID     string
	IssueNumber   int
	WrikeTaskID   string
	Hours         float64
	IsWrikeParent bool
	SubIssues     []int
}

// GraphQLRequest represents a GraphQL request.
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// GraphQLResponse represents a GraphQL response.
type GraphQLResponse struct {
	Data   map[string]interface{} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

// GetProjectItemByIssue retrieves project item data for a GitHub issue using GraphQL.
func (c *Client) GetProjectItemByIssue(issueNumber int, projectNumber int) (*ProjectItem, error) {
	// Split repo into owner and name
	owner, repo := c.splitRepo()

	query := `
		query($owner: String!, $repo: String!, $issueNumber: Int!, $projectNumber: Int!) {
			repository(owner: $owner, name: $repo) {
				issue(number: $issueNumber) {
					number
					projectItems(first: 10) {
						nodes {
							id
							project {
								id
								number
							}
							fieldValues(first: 20) {
								nodes {
									... on ProjectV2ItemFieldNumberValue {
										number
										field {
											... on ProjectV2Field {
												name
											}
										}
									}
									... on ProjectV2ItemFieldTextValue {
										text
										field {
											... on ProjectV2Field {
												name
											}
										}
									}
									... on ProjectV2ItemFieldSingleSelectValue {
										name
										field {
											... on ProjectV2SingleSelectField {
												name
											}
										}
									}
								}
							}
						}
					}
					body
				}
			}
		}
	`

	variables := map[string]interface{}{
		"owner":         owner,
		"repo":          repo,
		"issueNumber":   issueNumber,
		"projectNumber": projectNumber,
	}

	resp, err := c.executeGraphQL(query, variables)
	if err != nil {
		return nil, err
	}

	// Parse the response
	projectItem := &ProjectItem{
		IssueNumber: issueNumber,
	}

	// Navigate through the response structure
	if repo, ok := resp.Data["repository"].(map[string]interface{}); ok {
		if issue, ok := repo["issue"].(map[string]interface{}); ok {
			// Get issue body for subtask references
			if body, ok := issue["body"].(string); ok {
				subIssueMatches := subIssuesRegex.FindAllStringSubmatch(body, -1)
				for _, match := range subIssueMatches {
					if len(match) >= 2 {
						num, err := strconv.Atoi(match[1])
						if err != nil {
							continue
						}
						if num > 0 {
							projectItem.SubIssues = append(projectItem.SubIssues, num)
						}
					}
				}
			}

			// Get project items
			if projectItems, ok := issue["projectItems"].(map[string]interface{}); ok {
				if nodes, ok := projectItems["nodes"].([]interface{}); ok && len(nodes) > 0 {
					// Find the matching project
					for _, node := range nodes {
						if itemNode, ok := node.(map[string]interface{}); ok {
							if project, ok := itemNode["project"].(map[string]interface{}); ok {
								if projNum, ok := project["number"].(float64); ok && int(projNum) == projectNumber {
									if id, ok := itemNode["id"].(string); ok {
										projectItem.ID = id
									}
									if projID, ok := project["id"].(string); ok {
										projectItem.ProjectID = projID
									}

									// Parse field values
									if fieldValues, ok := itemNode["fieldValues"].(map[string]interface{}); ok {
										if fieldNodes, ok := fieldValues["nodes"].([]interface{}); ok {
											for _, fieldNode := range fieldNodes {
												if field, ok := fieldNode.(map[string]interface{}); ok {
													c.parseProjectField(field, projectItem)
												}
											}
										}
									}
									break
								}
							}
						}
					}
				}
			}
		}
	}

	return projectItem, nil
}

// parseProjectField parses a single project field value.
func (c *Client) parseProjectField(field map[string]interface{}, item *ProjectItem) {
	var fieldName string

	// Get field name from different field types
	if fieldInfo, ok := field["field"].(map[string]interface{}); ok {
		if name, ok := fieldInfo["name"].(string); ok {
			fieldName = name
		}
	}

	// Parse based on field name and type
	switch fieldName {
	case "Hours":
		if num, ok := field["number"].(float64); ok {
			item.Hours = num
		}
	case "Wrike Task ID":
		if text, ok := field["text"].(string); ok {
			item.WrikeTaskID = text
		}
	case "Wrike Parent":
		// Could be checkbox (name) or select field
		// Accept various boolean-like values (case-insensitive)
		if name, ok := field["name"].(string); ok {
			item.IsWrikeParent = parseBooleanValue(name)
		}
	}
}

// parseBooleanValue converts common boolean-like strings to bool (case-insensitive).
func parseBooleanValue(value string) bool {
	valueLower := strings.ToLower(strings.TrimSpace(value))
	return valueLower == "yes" || valueLower == "true" || valueLower == "enabled" || valueLower == "checked"
}

// UpdateProjectField updates a custom field value in GitHub Projects V2.
func (c *Client) UpdateProjectField(projectID string, itemID string, fieldID string, value interface{}) error {
	var mutation string

	// Determine the mutation based on value type
	switch value.(type) {
	case string:
		mutation = `
			mutation($projectId: ID!, $itemId: ID!, $fieldId: ID!, $value: String!) {
				updateProjectV2ItemFieldValue(
					input: {
						projectId: $projectId
						itemId: $itemId
						fieldId: $fieldId
						value: { text: $value }
					}
				) {
					projectV2Item {
						id
					}
				}
			}
		`
	case float64, int:
		mutation = `
			mutation($projectId: ID!, $itemId: ID!, $fieldId: ID!, $value: Float!) {
				updateProjectV2ItemFieldValue(
					input: {
						projectId: $projectId
						itemId: $itemId
						fieldId: $fieldId
						value: { number: $value }
					}
				) {
					projectV2Item {
						id
					}
				}
			}
		`
	default:
		return fmt.Errorf("unsupported field value type")
	}

	variables := map[string]interface{}{
		"projectId": projectID,
		"itemId":    itemID,
		"fieldId":   fieldID,
		"value":     value,
	}

	_, err := c.executeGraphQL(mutation, variables)
	return err
}

// GetProjectFieldID retrieves the field ID for a given field name.
func (c *Client) GetProjectFieldID(projectID string, fieldName string) (string, error) {
	query := `
		query($projectId: ID!) {
			node(id: $projectId) {
				... on ProjectV2 {
					fields(first: 20) {
						nodes {
							... on ProjectV2Field {
								id
								name
							}
							... on ProjectV2SingleSelectField {
								id
								name
							}
						}
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"projectId": projectID,
	}

	resp, err := c.executeGraphQL(query, variables)
	if err != nil {
		return "", err
	}

	// Parse response to find field ID
	if node, ok := resp.Data["node"].(map[string]interface{}); ok {
		if fields, ok := node["fields"].(map[string]interface{}); ok {
			if nodes, ok := fields["nodes"].([]interface{}); ok {
				for _, fieldNode := range nodes {
					if field, ok := fieldNode.(map[string]interface{}); ok {
						if name, ok := field["name"].(string); ok && name == fieldName {
							if id, ok := field["id"].(string); ok {
								return id, nil
							}
						}
					}
				}
			}
		}
	}

	return "", fmt.Errorf("field '%s' not found in project", fieldName)
}

// executeGraphQL executes a GraphQL query.
func (c *Client) executeGraphQL(query string, variables map[string]interface{}) (*GraphQLResponse, error) {
	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal GraphQL request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.botToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GraphQL request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GraphQL API error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var graphQLResp GraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&graphQLResp); err != nil {
		return nil, fmt.Errorf("failed to decode GraphQL response: %w", err)
	}

	if len(graphQLResp.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL errors: %v", graphQLResp.Errors)
	}

	return &graphQLResp, nil
}
