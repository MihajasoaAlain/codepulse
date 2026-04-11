package server

import (
	"bytes"
	"codepulse/internal/features/users/handlers"
	"codepulse/internal/features/users/models"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeUserRepository struct{}

func (f fakeUserRepository) Create(ctx context.Context, user *models.User) error {
	return nil
}

func TestHelloWorldHandler(t *testing.T) {
	s := &Server{}
	r := gin.New()
	r.GET("/", s.HelloWorldHandler)
	// Create a test HTTP request
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	// Serve the HTTP request
	r.ServeHTTP(rr, req)
	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	// Check the response body
	expected := "{\"message\":\"Hello World\"}"
	if rr.Body.String() != expected {
		t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestUsersRouteRegistered(t *testing.T) {
	handlers.SetUserRepository(fakeUserRepository{})
	s := &Server{}
	r := s.RegisterRoutes()

	payload := []byte(`{"username":"bob","email":"bob@example.com"}`)
	req, err := http.NewRequest(http.MethodPost, "/users", bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v", rr.Code, http.StatusCreated)
	}
}
