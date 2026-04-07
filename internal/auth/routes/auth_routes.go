package routes

import (
	"codepulse/internal/auth/handlers"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(r *gin.Engine) {

	auth := r.Group("/auth")

	auth.GET("/github/login", handlers.GithubLogin)

	auth.GET("/github/callback", handlers.GithubCallback)
}
