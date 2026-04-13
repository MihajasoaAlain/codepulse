package routes

import (
	githubhandlers "codepulse/internal/features/github/handlers"

	"github.com/gin-gonic/gin"
)

func GithubRoutes(r *gin.Engine) {
	github := r.Group("/github")
	github.GET("/commits/:email", githubhandlers.GetCommit)
}
