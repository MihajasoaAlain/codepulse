package config

import (
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var GithubOAuthConfig = &oauth2.Config{
	ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
	ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
	RedirectURL:  os.Getenv("GITHUB_REDIRECT_URL"),
	Scopes:       []string{"user:email"},
	Endpoint:     github.Endpoint,
}
