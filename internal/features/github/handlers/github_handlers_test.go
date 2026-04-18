package handlers

import (
	"codepulse/internal/features/users/dto"
	usershandlers "codepulse/internal/features/users/handlers"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"codepulse/internal/features/users/models"

	"github.com/gin-gonic/gin"
)

type fakeCommitUserRepository struct{}

func (f fakeCommitUserRepository) Create(_ context.Context, _ *models.User) error {
	return nil
}

func (f fakeCommitUserRepository) FindByEmail(_ context.Context, _ string) (*models.User, error) {
	return &models.User{}, nil
}

func (f fakeCommitUserRepository) GetGithubToken(_ context.Context, _ dto.UserRequest) (dto.UserGithubTokenResponse, error) {
	return dto.UserGithubTokenResponse{GithubToken: "test-token"}, nil
}

func TestGetCommit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	usershandlers.SetUserRepository(fakeCommitUserRepository{})

	githubServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/user":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"login":"octocat"}`))
		case "/search/commits":
			if got := r.Header.Get("Accept"); got != "application/vnd.github.cloak-preview+json" {
				t.Fatalf("unexpected accept header: %s", got)
			}
			if !strings.Contains(r.URL.RawQuery, "author:octocat") {
				t.Fatalf("expected author filter in query, got %s", r.URL.RawQuery)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"total_count":3,"items":[]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer githubServer.Close()

	oldBaseURL := githubAPIBaseURL
	githubAPIBaseURL = githubServer.URL
	defer func() { githubAPIBaseURL = oldBaseURL }()

	r := gin.New()
	r.GET("/github/commits/:email", GetCommit)

	req, err := http.NewRequest(http.MethodGet, "/github/commits/alice%40example.com?commitDate=2026-01-01T00:00:00Z", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}

	if body["committer"] != "octocat" {
		t.Fatalf("expected committer octocat, got %v", body["committer"])
	}

	if body["commitCount"] != float64(3) {
		t.Fatalf("expected commitCount 3, got %v", body["commitCount"])
	}

	if !strings.Contains(body["commitDate"].(string), "2026-01-01") {
		t.Fatalf("unexpected commitDate: %v", body["commitDate"])
	}

	if parsed, _ := url.QueryUnescape(req.URL.RawQuery); !strings.Contains(parsed, "commitDate") {
		t.Fatalf("expected commitDate query")
	}
}
