package dto

import "go.mongodb.org/mongo-driver/bson/primitive"

type GithubRepo struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
}
type GithubCommit struct {
	CommitDate  primitive.DateTime `json:"commitDate"`
	Committer   string             `json:"committer"`
	CommitCount int                `json:"commitCount"`
}
