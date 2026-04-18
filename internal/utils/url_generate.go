package utils

import "codepulse/internal/features/github/constant"

import "go.mongodb.org/mongo-driver/bson/primitive"

func GenerateCommitsURL(username string, commitDate primitive.DateTime) string {
	return constant.GithubApiUrl + "/search/commits?q=author:" + username + "/commits?since=" + commitDate.Time().Format("2006-01-02T15:04:05Z07:00")
}
