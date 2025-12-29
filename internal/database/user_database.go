package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/neilsmahajan/productivity-timer/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *service) getUsersCollection() *mongo.Collection {
	return s.db.Database(database).Collection("users")
}

func (s *service) FindOrCreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	collection := s.getUsersCollection()

	// Try to find existing user by provider and provider ID
	filter := bson.M{
		"provider":    user.Provider,
		"provider_id": user.ProviderID,
	}

	var existingUser models.User
	err := collection.FindOne(ctx, filter).Decode(&existingUser)

	if err == nil {
		// User exists, update last login
		update := bson.M{
			"$set": bson.M{
				"last_login_at": time.Now(),
				"email":         user.Email,
				"name":          user.Name,
				"first_name":    user.FirstName,
				"last_name":     user.LastName,
				"nick_name":     user.NickName,
				"avatar_url":    user.AvatarURL,
			},
		}
		_, err = collection.UpdateOne(ctx, filter, update)
		if err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
		existingUser.LastLoginAt = time.Now()
		return &existingUser, nil
	}

	if !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// User doesn't exist, create new one
	now := time.Now()
	user.CreatedAt = now
	user.LastLoginAt = now

	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *service) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	collection := s.getUsersCollection()

	var user models.User
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}
