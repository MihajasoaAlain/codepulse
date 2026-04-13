package dto

type CommitRequest struct {
	Email      string `uri:"email" form:"email" json:"email"`
	CommitDate string `form:"commitDate" json:"commitDate"`
}
