package database

import (
	"context"
	"errors"

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

func (s *service) FindTimerSession(ctx context.Context, userId, tag string) (*models.TimerSession, error) {
	collection := s.getTimerSessionsCollection()
	var timerSession models.TimerSession
	err := collection.FindOne(ctx, bson.M{"user_id": userId, "tag": tag}).Decode(&timerSession)
	if err != nil {
		return nil, err
	}
	return &timerSession, nil
}

func (s *service) FindOrCreateTimerSession(ctx context.Context, userId, tag string) (*models.TimerSession, error) {
	timerSession, err := s.FindTimerSession(ctx, userId, tag)
	if errors.Is(err, mongo.ErrNoDocuments) {
		timerSession = models.NewTimerSession(userId, tag)
		if err = s.CreateTimerSession(ctx, timerSession); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return timerSession, nil
}
