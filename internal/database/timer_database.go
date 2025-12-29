package database

import (
	"context"

	"github.com/neilsmahajan/productivity-timer/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *service) getTimerSessionsCollection() *mongo.Collection {
	return s.db.Database(database).Collection("timers")
}

func (s *service) UpdateTimerSession(ctx context.Context, timerSession *models.TimerSession) error {
	collection := s.getTimerSessionsCollection()
	filter := bson.M{"_id": timerSession.ID}

	if _, err := collection.UpdateOne(ctx, filter, bson.M{"$set": timerSession}); err != nil {
		return err
	}

	return nil
}

func (s *service) CreateTimerSession(ctx context.Context, timerSession *models.TimerSession) error {
	collection := s.getTimerSessionsCollection()
	if _, err := collection.InsertOne(ctx, timerSession); err != nil {
		return err
	}

	return nil
}

func (s *service) GetTimerSession(ctx context.Context, userId, tag string) (*models.TimerSession, error) {
	collection := s.getTimerSessionsCollection()
	var timerSession models.TimerSession
	err := collection.FindOne(ctx, bson.M{"user_id": userId, "tag": tag}).Decode(&timerSession)
	if err != nil {
		return nil, err
	}
	return &timerSession, nil
}
