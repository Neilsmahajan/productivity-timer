package database

import (
	"context"
	"fmt"
	"time"

	"github.com/neilsmahajan/productivity-timer/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *service) getTagStatsCollection() *mongo.Collection {
	return s.db.Database(database).Collection("tagstats")
}

func (s *service) UpdateTagStats(ctx context.Context, userTagStats *models.UserTagStats, elapsed int64) error {
	collection := s.getTagStatsCollection()
	update := bson.M{
		"$set": bson.M{
			"total_duration": userTagStats.TotalDuration + elapsed,
			"last_updated":   time.Now(),
			"session_count":  userTagStats.SessionCount + 1,
		},
	}

	if _, err := collection.UpdateOne(ctx, bson.M{"id": userTagStats.ID}, update); err != nil {
		return err
	}

	return nil
}

func (s *service) FindOrCreateTagStats(ctx context.Context, userId string, tag string) (*models.UserTagStats, error) {
	collection := s.getTagStatsCollection()

	filter := bson.M{
		"user_id": userId,
		"tag":     tag,
	}

	var existingTagStats models.UserTagStats
	err := collection.FindOne(ctx, filter).Decode(&existingTagStats)
	if err == nil {
		update := bson.M{
			"$set": bson.M{
				"last_updated": time.Now(),
			},
		}
		_, err = collection.UpdateOne(ctx, filter, update)
		if err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
		return &existingTagStats, nil
	}

	newTagStats := models.NewUserTagStats(userId, tag)
	_, err = collection.InsertOne(ctx, newTagStats)
	if err != nil {
		return nil, fmt.Errorf("failed to insert new tag stats: %w", err)
	}

	return newTagStats, nil
}
