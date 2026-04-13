package dto

type GithubRepo struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
}
type GithubCommit struct {
	CommitDate  string `json:"commitDate"`
	Committer   string `json:"committer"`
	CommitCount int    `json:"commitCount"`
}
