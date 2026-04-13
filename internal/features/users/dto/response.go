package dto

type addUserErrorResponse struct {
	Error string `json:"error"`
}
type UserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	CreatedAt string `json:"createdAt"`
}

type UserGithubTokenResponse struct {
	Username           string `json:"username"`
	GithubToken        string `json:"githubToken"`
	GithubRefreshToken string `json:"githubRefreshToken"`
}
