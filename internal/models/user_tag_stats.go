package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Period string

const (
	PeriodDaily   Period = "daily"
	PeriodWeekly  Period = "weekly"
	PeriodMonthly Period = "monthly"
	PeriodCustom  Period = "custom"
)

type UserTagStats struct {
	ID            primitive.ObjectID `bson:"_id" json:"id"`
	UserID        string             `bson:"user_id" json:"userId"`
	Tag           string             `bson:"tag" json:"tag"`
	TotalDuration int64              `bson:"total_duration" json:"totalDuration"` // in seconds
	SessionCount  int                `bson:"session_count" json:"sessionCount"`
	LastUpdated   time.Time          `bson:"last_updated" json:"lastUpdated"`
}

func NewUserTagStats(userID, tag string) *UserTagStats {
	return &UserTagStats{
		ID:            primitive.NewObjectID(),
		UserID:        userID,
		Tag:           tag,
		TotalDuration: 0,
		SessionCount:  1,
		LastUpdated:   time.Now(),
	}
}
