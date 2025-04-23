package user

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// User is a struct holding the necessary data to sign up any user to the WrikeMeUp service.
type User struct {
	GitHubUsername string `json:"github_username"`
	WrikeEmail     string `json:"wrike_email"`
	WrikeToken     string `json:"wrike_token"`
}

// DecodeUserFromEnv gets the user credentials from the environment variable and decodes it.
func DecodeUserFromEnv(gitHubUsername string, encodedUserCredentials string) (*User, error) {
	usersJson, err := base64.StdEncoding.DecodeString(encodedUserCredentials)
	if err != nil {
		return nil, fmt.Errorf("user: failed to decode encoded user credentials: %w", err)
	}
	var users []*User
	if err := json.Unmarshal(usersJson, &users); err != nil {
		return nil, fmt.Errorf("user: failed to unmarshal users: %w", err)
	}
	for _, u := range users {
		if u.GitHubUsername == gitHubUsername {
			return u, nil
		}
	}
	return nil, fmt.Errorf("wrikemeup: no credentials found for GitHub user: %s", gitHubUsername)
}
