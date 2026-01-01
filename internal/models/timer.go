package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TimerStatus string

const (
	StatusRunning   TimerStatus = "running"
	StatusStopped   TimerStatus = "stopped"
	StatusCompleted TimerStatus = "completed"
)

type TimerSession struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	UserID      string             `bson:"user_id" json:"userId"`
	Tag         string             `bson:"tag" json:"tag"`
	StartTime   time.Time          `bson:"start_time" json:"startTime"`
	EndTime     *time.Time         `bson:"end_time,omitempty" json:"endTime,omitempty"`
	Duration    int64              `bson:"duration" json:"duration"` // Duration in seconds
	Status      TimerStatus        `bson:"status" json:"status"`     // e.g., "running", "stopped"
	CreatedAt   time.Time          `bson:"created_at" json:"createdAt"`
	LastUpdated time.Time          `bson:"last_updated" json:"lastUpdated"`
}

func NewTimerSession(userID, tag string) *TimerSession {
	return &TimerSession{
		ID:          primitive.NewObjectID(),
		UserID:      userID,
		Tag:         tag,
		StartTime:   time.Now(),
		Duration:    0,
		Status:      StatusRunning,
		CreatedAt:   time.Now(),
		LastUpdated: time.Now(),
	}
}
