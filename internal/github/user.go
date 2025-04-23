package github

import (
	"fmt"

	userpkg "github.com/lopezator/wrikemeup/internal/user"
)

func test(gitHubUsername string, users []*userpkg.User) (*userpkg.User, error) {
	var user *userpkg.User
	for _, u := range users {
		if u.GitHubUsername == gitHubUsername {
			user = u
		}
	}
	if user == nil {
		return nil, fmt.Errorf("wrikemeup: no credentials found for GitHub user: %s", gitHubUsername)
	}
	return user, nil
}
