package repository

import (
	"codepulse/internal/features/users/dto"
	"codepulse/internal/features/users/models"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	GetGithubToken(ctx context.Context, req dto.UserRequest) (dto.UserGithubTokenResponse, error)
}

type MongoUserRepository struct {
	collection *mongo.Collection
}

func NewMongoUserRepository() *MongoUserRepository {
	host := os.Getenv("BLUEPRINT_DB_HOST")
	port := os.Getenv("BLUEPRINT_DB_PORT")
	database := os.Getenv("BLUEPRINT_DB_DATABASE")
	if database == "" {
		database = "codepulse"
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s", host, port)))
	if err != nil {
		log.Printf("users repository: mongo connect failed: %v", err)
		return nil
	}

	return &MongoUserRepository{
		collection: client.Database(database).Collection("users"),
	}
}

func (r *MongoUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	if r == nil || r.collection == nil {
		return nil, errors.New("users repository unavailable")
	}

	var existing models.User
	err := r.collection.FindOne(ctx, models.User{Email: email}).Decode(&existing)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	fmt.Println(existing)
	return &existing, nil
}

func (r *MongoUserRepository) Create(ctx context.Context, user *models.User) error {
	if r == nil || r.collection == nil {
		return errors.New("users repository unavailable")
	}

	_, err := r.collection.InsertOne(ctx, user)
	return err
}
func (r *MongoUserRepository) GetGithubToken(ctx context.Context, req dto.UserRequest) (dto.UserGithubTokenResponse, error) {
	if r == nil || r.collection == nil {
		return dto.UserGithubTokenResponse{}, errors.New("users repository unavailable")
	}

	var user models.User
	err := r.collection.FindOne(
		ctx,
		bson.M{"email": strings.TrimSpace(req.Email)},
		options.FindOne().SetProjection(bson.M{
			"githubToken":        1,
			"githubRefreshToken": 1,
			"username":           1,
			"_id":                0,
		}),
	).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return dto.UserGithubTokenResponse{}, ErrUserNotFound
		}
		return dto.UserGithubTokenResponse{}, err
	}

	return dto.UserGithubTokenResponse{
		Username:           user.Username,
		GithubToken:        user.GithubToken,
		GithubRefreshToken: user.GithubRefreshToken,
	}, nil
}
