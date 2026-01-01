package database

import (
	"context"
	"fmt"
	"log"

	"github.com/neilsmahajan/productivity-timer/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *service) getUserTagStatsCollection() *mongo.Collection {
	return s.db.Database(database).Collection("tagstats")
}

func (s *service) UpdateUserTagStats(ctx context.Context, userTagStats *models.UserTagStats) error {
	collection := s.getUserTagStatsCollection()
	filter := bson.M{"_id": userTagStats.ID}

	if _, err := collection.UpdateOne(ctx, filter, bson.M{"$set": userTagStats}); err != nil {
		return err
	}

	return nil
}

func (s *service) CreateUserTagStats(ctx context.Context, userTagStats *models.UserTagStats) error {
	collection := s.getUserTagStatsCollection()
	_, err := collection.InsertOne(ctx, userTagStats)
	if err != nil {
		return fmt.Errorf("failed to insert new tag stats: %w", err)
	}
	return nil
}

func (s *service) FindUserTagStats(ctx context.Context, userId, tag string) (*models.UserTagStats, error) {
	collection := s.getUserTagStatsCollection()
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

func (s *service) FindAllUserTagStats(ctx context.Context, userId string) ([]*models.UserTagStats, error) {
	collection := s.getUserTagStatsCollection()
	filter := bson.M{
		"user_id": userId,
	}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err = cursor.Close(ctx)
		if err != nil {
			log.Println(err)
		}
	}(cursor, ctx)
	var tagStats []*models.UserTagStats
	if err = cursor.All(ctx, &tagStats); err != nil {
		return nil, err
	}
	return tagStats, nil
}

func (s *service) DeleteUserTagStats(ctx context.Context, userId, tag string) error {
	collection := s.getUserTagStatsCollection()

	filter := bson.M{"user_id": userId, "tag": tag}

	if _, err := collection.DeleteOne(ctx, filter); err != nil {
		return err
	}
	return nil
}
