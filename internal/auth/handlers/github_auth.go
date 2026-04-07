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

type githubAuthErrorResponse struct {
	Error string `json:"error"`
}

type githubCallbackResponse struct {
	User  map[string]interface{} `json:"user"`
	Token string                 `json:"token"`
}

// GithubLogin godoc
// @Summary Start GitHub OAuth login
// @Description Redirects the client to GitHub's OAuth consent page.
// @Tags Auth
// @Success 307 {string} string "Temporary Redirect to GitHub"
// @Router /auth/github/login [get]
func GithubLogin(c *gin.Context) {

	url := config.GithubOAuthConfig.AuthCodeURL("state")

	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GithubCallback godoc
// @Summary Handle GitHub OAuth callback
// @Description Exchanges the GitHub OAuth code, fetches user profile, and returns a JWT.
// @Tags Auth
// @Param code query string true "GitHub OAuth authorization code"
// @Produce json
// @Success 200 {object} githubCallbackResponse
// @Failure 500 {object} githubAuthErrorResponse
// @Router /auth/github/callback [get]
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
		"token": "Bearer " + jwtToken,
	})
}
