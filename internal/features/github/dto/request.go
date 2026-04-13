package dto

import "go.mongodb.org/mongo-driver/bson/primitive"

type CommitRequest struct {
	commitDate primitive.DateTime `json:"commitDate,omitempty" bson:"commitDate,omitempty"`
}
