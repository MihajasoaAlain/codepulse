package handlers

import (
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"codepulse/internal/features/users/models"
	"codepulse/internal/features/users/repository"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"codepulse/internal/features/users/dto"
)

var (
	userRepo   repository.UserRepository
	userRepoMu sync.RWMutex
)

func SetUserRepository(repo repository.UserRepository) {
	userRepoMu.Lock()
	defer userRepoMu.Unlock()
	userRepo = repo
}

func GetUserRepository() repository.UserRepository {
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

	repo := GetUserRepository()
	if repo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "users repository unavailable"})
		return
	}

	if err := repo.Create(c.Request.Context(), &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}
	response := dto.UserResponse{
		ID:        user.ID.Hex(),
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.String(),
	}

	c.JSON(http.StatusCreated, response)
}

// FindByEmail godoc
// @Summary Find user by email
// @Description Find a user by email.
// @Tags Users
// @Accept json
// @Produce json
// @Param email path string true "User email"
// @Success 200 {object} models.User
// @Router  /users/{email} [get]
func FindByEmail(c *gin.Context) {
	email := c.Param("email")
	repo := GetUserRepository()

	if repo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "users repository unavailable"})
		return
	}

	user, err := repo.FindByEmail(c.Request.Context(), email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find user"})
		return
	}
	response := dto.UserResponse{
		ID:        user.ID.Hex(),
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.String(),
	}

	c.JSON(http.StatusOK, response)
}

// GetGithubToken godoc
// @Summary Get GitHub token
// @Description Get the GitHub token for a user.
// @Tags Users
// @Accept json
// @Produce json
// @Param email path string true "User email"
// @Success 200 {object} dto.UserGithubTokenResponse
// @Router /users/{email}/github [get]
func GetGithubToken(c *gin.Context) {
	repo := GetUserRepository()

	if repo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "users repository unavailable"})
		return
	}

	response, err := repo.GetGithubToken(c.Request.Context(), dto.UserRequest{
		Email: c.Param("email"),
	})
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get github token"})
		return
	}

	c.JSON(http.StatusOK, response)
}
