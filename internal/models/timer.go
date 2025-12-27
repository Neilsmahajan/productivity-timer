package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TimerSession struct {
	ID        primitive.ObjectID `bson:"_id" json:"id"`
	UserID    string             `bson:"user_id" json:"userId"`
	Tag       string             `bson:"tag" json:"tag"`
	StartTime time.Time          `bson:"start_time" json:"startTime"`
	EndTime   *time.Time         `bson:"end_time,omitempty" json:"endTime,omitempty"`
	Duration  int64              `bson:"duration" json:"duration"` // Duration in seconds
	Status    string             `bson:"status" json:"status"`     // e.g., "running", "stopped"
	CreatedAt time.Time          `bson:"created_at" json:"createdAt"`
}

func NewTimerSession(userID, tag string) *TimerSession {
	return &TimerSession{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		Tag:       tag,
		StartTime: time.Now(),
		Status:    "running",
		CreatedAt: time.Now(),
	}
}
