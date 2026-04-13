package routes

import (
	"codepulse/internal/features/users/handlers"

	"github.com/gin-gonic/gin"
)

func UsersRoutes(r *gin.Engine) {
	users := r.Group("/users")
	users.POST("", handlers.AddUser)
	users.GET("/:email", handlers.FindByEmail)
	users.GET("/:email/github-token", handlers.GetGithubToken)
}
