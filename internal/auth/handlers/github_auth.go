package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"codepulse/internal/auth/config"
	"codepulse/internal/features/users/models"
	"codepulse/internal/features/users/repository"
	"codepulse/internal/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type githubAuthErrorResponse struct {
	Error string `json:"error"`
}

type githubCallbackResponse struct {
	User  interface{} `json:"user"`
	Token string      `json:"token"`
}

var (
	userRepo   repository.UserRepository
	userRepoMu sync.RWMutex
)

func getUserRepository() repository.UserRepository {
	userRepoMu.RLock()
	repo := userRepo
	userRepoMu.RUnlock()

	if repo != nil {
		return repo
	}

	userRepoMu.Lock()
	defer userRepoMu.Unlock()

	if userRepo == nil {
		userRepo = repository.NewMongoUserRepository()
	}

	return userRepo
}

// GithubLogin godoc
// @Summary Start GitHub OAuth login
// @Tags Auth
// @Success 307 {string} string "Redirect to GitHub"
// @Router /auth/github/login [get]
func GithubLogin(c *gin.Context) {

	url := config.GithubOAuthConfig.AuthCodeURL("state")

	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GithubCallback godoc
// @Summary Handle GitHub OAuth callback
// @Tags Auth
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "token exchange failed",
		})
		return
	}

	client := config.GithubOAuthConfig.Client(
		context.Background(),
		token,
	)

	// récupérer profil GitHub
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch github user",
		})
		return
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var githubUser map[string]interface{}

	if err := json.Unmarshal(body, &githubUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "invalid github response",
		})
		return
	}

	username := ""
	if login, ok := githubUser["login"].(string); ok {
		username = strings.TrimSpace(login)
	}

	email := ""

	emailResp, err := client.Get("https://api.github.com/user/emails")
	if err == nil {

		defer emailResp.Body.Close()

		emailBody, _ := io.ReadAll(emailResp.Body)

		var emails []map[string]interface{}

		json.Unmarshal(emailBody, &emails)

		for _, e := range emails {

			if primary, ok := e["primary"].(bool); ok && primary {

				if em, ok := e["email"].(string); ok {

					email = strings.TrimSpace(em)

					break
				}
			}
		}
	}

	jwtToken, err := utils.GenerateJWT(username)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "JWT generation failed",
		})
		return
	}

	now := time.Now().UTC()

	user := models.User{
		ID:        primitive.NewObjectID(),
		Username:  username,
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}

	repo := getUserRepository()

	if repo == nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "user repository unavailable",
		})

		return
	}
	fmt.Println("GitHub user:", githubUser)
	fmt.Println(email)
	if email != "" {
		existingUser, err := repo.FindByEmail(c.Request.Context(), email)
		fmt.Println(existingUser)
		if err == nil && existingUser != nil {
			fmt.Println("user already exists")
			c.JSON(http.StatusOK, gin.H{
				"user":  existingUser,
				"token": "Bearer " + jwtToken,
			})
			return
		}
		if err != nil && err != repository.ErrUserNotFound {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to check existing user",
			})
			return
		}
	}

	if err := repo.Create(c.Request.Context(), &user); err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create user",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{

		"user":  user,
		"token": "Bearer " + jwtToken,
	})
}
