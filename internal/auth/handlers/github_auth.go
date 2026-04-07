package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"codepulse/internal/auth/config"
	"codepulse/internal/utils"
)

func GithubLogin(c *gin.Context) {

	url := config.GithubOAuthConfig.AuthCodeURL("state")

	c.Redirect(http.StatusTemporaryRedirect, url)
}

func GithubCallback(c *gin.Context) {

	code := c.Query("code")

	token, err := config.GithubOAuthConfig.Exchange(
		context.Background(),
		code,
	)

	if err != nil {

		c.JSON(500, gin.H{"error": "token exchange failed"})

		return
	}

	client := config.GithubOAuthConfig.Client(
		context.Background(),
		token,
	)

	resp, err := client.Get("https://api.github.com/user")

	if err != nil {

		c.JSON(500, gin.H{"error": "failed to get user"})

		return
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var githubUser map[string]interface{}

	json.Unmarshal(body, &githubUser)

	jwtToken, err := utils.GenerateJWT(
		githubUser["login"].(string),
	)

	if err != nil {

		c.JSON(500, gin.H{"error": "JWT error"})

		return
	}

	c.JSON(200, gin.H{
		"user":  githubUser,
		"token": jwtToken,
	})
}
