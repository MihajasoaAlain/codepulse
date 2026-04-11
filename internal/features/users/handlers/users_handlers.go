package handlers

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"codepulse/internal/features/users/models"
	"codepulse/internal/features/users/repository"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type addUserErrorResponse struct {
	Error string `json:"error"`
}

var (
	userRepo   repository.UserRepository
	userRepoMu sync.RWMutex
)

func SetUserRepository(repo repository.UserRepository) {
	userRepoMu.Lock()
	defer userRepoMu.Unlock()
	userRepo = repo
}

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

// AddUser godoc
// @Summary Create user
// @Description Create a new user from username and email.
// @Tags Users
// @Accept json
// @Produce json
// @Param user body models.CreateUserRequest true "User payload"
// @Success 201 {object} models.User
// @Failure 400 {object} addUserErrorResponse
// @Failure 500 {object} addUserErrorResponse
// @Router /users [post]
func AddUser(c *gin.Context) {
	var req models.CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Email) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and email are required"})
		return
	}

	now := time.Now().UTC()
	user := models.User{
		ID:        primitive.NewObjectID(),
		Username:  strings.TrimSpace(req.Username),
		Email:     strings.TrimSpace(req.Email),
		CreatedAt: now,
		UpdatedAt: now,
	}

	repo := getUserRepository()
	if repo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "users repository unavailable"})
		return
	}

	if err := repo.Create(c.Request.Context(), &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func FindByEmail(c *gin.Context) (models.User, error) {
	email := c.Param("email")
	repo := getUserRepository()

	if repo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "users repository unavailable"})
		return models.User{}, nil
	}

	user, err := repo.FindByEmail(c.Request.Context(), email)
	if err != nil {
		return models.User{}, err
	}

	return *user, nil
}
