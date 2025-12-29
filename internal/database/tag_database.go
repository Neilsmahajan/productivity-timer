package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/neilsmahajan/productivity-timer/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *service) getTagStatsCollection() *mongo.Collection {
	return s.db.Database(database).Collection("tagstats")
}

func (s *service) UpdateTagStats(ctx context.Context, userTagStats *models.UserTagStats) error {
	collection := s.getTagStatsCollection()
	filter := bson.M{"_id": userTagStats.ID}
	update := bson.M{"$set": bson.M{}}

	if _, err := collection.UpdateOne(ctx, filter, update); err != nil {
		return err
	}

	return nil
}

func (s *service) CreateTagStats(ctx context.Context, userTagStats *models.UserTagStats) error {
	collection := s.getTagStatsCollection()
	_, err := collection.InsertOne(ctx, userTagStats)
	if err != nil {
		return fmt.Errorf("failed to insert new tag stats: %w", err)
	}
	return nil
}

func (s *service) FindTagStats(ctx context.Context, userId, tag string) (*models.UserTagStats, error) {
	collection := s.getTagStatsCollection()
	filter := bson.M{
		"user_id": userId,
		"tag":     tag,
	}

	var existingTagStats models.UserTagStats
	if err := collection.FindOne(ctx, filter).Decode(&existingTagStats); err != nil {
		return nil, err
	}

	return &existingTagStats, nil
}

func (s *service) FindOrCreateTagStats(ctx context.Context, userID, tag string) (*models.UserTagStats, error) {
	userTagStats, err := s.FindTagStats(ctx, userID, tag)
	if errors.Is(err, mongo.ErrNoDocuments) {
		userTagStats = models.NewUserTagStats(userID, tag)
		if err = s.CreateTagStats(ctx, userTagStats); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return userTagStats, nil
}
