package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"codepulse/internal/features/users/dto"
	"codepulse/internal/features/users/models"

	"github.com/gin-gonic/gin"
)

type fakeUserRepository struct {
	createFn func(ctx context.Context, user *models.User) error
}

func (f fakeUserRepository) Create(ctx context.Context, user *models.User) error {
	if f.createFn != nil {
		return f.createFn(ctx, user)
	}
	return nil
}

func (f fakeUserRepository) FindByEmail(_ context.Context, _ string) (*models.User, error) {
	return nil, nil
}

func (f fakeUserRepository) GetGithubToken(_ context.Context, _ dto.UserRequest) (dto.UserGithubTokenResponse, error) {
	return dto.UserGithubTokenResponse{}, nil
}

func TestAddUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("creates user", func(t *testing.T) {
		SetUserRepository(fakeUserRepository{})
		r := gin.New()
		r.POST("/users", AddUser)

		payload := []byte(`{"username":"alice","email":"alice@example.com"}`)
		req, err := http.NewRequest(http.MethodPost, "/users", bytes.NewReader(payload))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
		}

		var body map[string]interface{}
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("invalid json response: %v", err)
		}

		if body["username"] != "alice" {
			t.Fatalf("expected username alice, got %v", body["username"])
		}
		if body["email"] != "alice@example.com" {
			t.Fatalf("expected email alice@example.com, got %v", body["email"])
		}
	})

	t.Run("rejects empty fields", func(t *testing.T) {
		SetUserRepository(fakeUserRepository{})
		r := gin.New()
		r.POST("/users", AddUser)

		payload := []byte(`{"username":" ","email":""}`)
		req, err := http.NewRequest(http.MethodPost, "/users", bytes.NewReader(payload))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("returns internal error when repository insert fails", func(t *testing.T) {
		SetUserRepository(fakeUserRepository{
			createFn: func(ctx context.Context, user *models.User) error {
				return errors.New("db insert failed")
			},
		})

		r := gin.New()
		r.POST("/users", AddUser)

		payload := []byte(`{"username":"alice","email":"alice@example.com"}`)
		req, err := http.NewRequest(http.MethodPost, "/users", bytes.NewReader(payload))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
		}
	})
}
