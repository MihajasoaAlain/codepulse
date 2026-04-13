package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"codepulse/internal/auth/config"
	"codepulse/internal/features/github/constant"
	githubdto "codepulse/internal/features/github/dto"
	usersdto "codepulse/internal/features/users/dto"
	usershandlers "codepulse/internal/features/users/handlers"
	"codepulse/internal/features/users/repository"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

var githubAPIBaseURL = constant.GithubApiUrl

type githubCommitSearchResponse struct {
	TotalCount int `json:"total_count"`
}

type githubUserResponse struct {
	Login string `json:"login"`
}

// GetCommit godoc
// @Summary Get GitHub commit count
// @Description Get the number of commits made by a user since a specified date.
// @Tags GitHub
// @Accept json
// @Produce json
// @Param email path string true "User email"
// @Param commitDate query string true "Commit date in RFC3339 format"
// @Success 200 {object} dto.GithubCommit
// @Router /github/commits/{email} [get]
func GetCommit(c *gin.Context) {
	repo := usershandlers.GetUserRepository()
	if repo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "users repository unavailable"})
		return
	}

	var req githubdto.CommitRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request parameters"})
		return
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request parameters"})
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.CommitDate = strings.TrimSpace(req.CommitDate)
	if req.Email == "" || req.CommitDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and commitDate are required"})
		return
	}

	since, err := time.Parse(time.RFC3339, req.CommitDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid commitDate format"})
		return
	}

	tokenResponse, err := repo.GetGithubToken(c.Request.Context(), usersdto.UserRequest{Email: req.Email})
	if err != nil {
		if strings.Contains(err.Error(), repository.ErrUserNotFound.Error()) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get github token"})
		return
	}

	if strings.TrimSpace(tokenResponse.GithubToken) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "github token unavailable"})
		return
	}

	client := config.GithubOAuthConfig.Client(c.Request.Context(), &oauth2.Token{
		AccessToken: tokenResponse.GithubToken,
		TokenType:   "Bearer",
	})

	totalCount, err := searchGithubCommits(c.Request.Context(), client, tokenResponse.Username, since)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch github commits"})
		return
	}

	c.JSON(http.StatusOK, githubdto.GithubCommit{
		CommitDate:  since.UTC().Format(time.RFC3339),
		Committer:   tokenResponse.Username,
		CommitCount: totalCount,
	})
}

func fetchGithubLogin(ctx context.Context, client *http.Client) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubAPIBaseURL+"/user", nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("github /user request failed: %s", strings.TrimSpace(string(body)))
	}

	var payload githubUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}

	if strings.TrimSpace(payload.Login) == "" {
		return "", fmt.Errorf("github login missing")
	}

	return strings.TrimSpace(payload.Login), nil
}

func searchGithubCommits(ctx context.Context, client *http.Client, login string, since time.Time) (int, error) {
	endpoint, err := url.Parse(githubAPIBaseURL + "/search/commits")
	if err != nil {
		return 0, err
	}

	query := endpoint.Query()
	query.Set("q", fmt.Sprintf("author:%s committer-date:%s", login, since.UTC().Format("2006-01-02")))
	query.Set("sort", "committer-date")
	query.Set("order", "desc")
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Accept", "application/vnd.github.cloak-preview+json")

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("github commits search failed: %s", strings.TrimSpace(string(body)))
	}

	var payload githubCommitSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return 0, err
	}

	return payload.TotalCount, nil
}
