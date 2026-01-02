package database

import (
	"context"
	"time"

	"github.com/neilsmahajan/productivity-timer/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func (s *service) FindTimerSession(ctx context.Context, userId, tag string, status models.TimerStatus) (*models.TimerSession, error) {
	collection := s.getTimerSessionsCollection()
	filter := bson.M{"user_id": userId, "tag": tag, "status": status}
	var timerSession models.TimerSession
	err := collection.FindOne(ctx, filter).Decode(&timerSession)
	if err != nil {
		return nil, err
	}
	return &timerSession, nil
}

// AbandonRunningTimers marks any running timers for a user+tag as completed.
// This handles orphaned timers when a user closes the tab while a timer is running.
func (s *service) AbandonRunningTimers(ctx context.Context, userId, tag string) error {
	collection := s.getTimerSessionsCollection()
	filter := bson.M{"user_id": userId, "tag": tag, "status": models.StatusRunning}
	update := bson.M{"$set": bson.M{
		"status":       models.StatusCompleted,
		"last_updated": time.Now(),
	}}

	_, err := collection.UpdateMany(ctx, filter, update)
	return err
}

// GetStatsSummary aggregates timer sessions for a user within a time period
func (s *service) GetStatsSummary(ctx context.Context, userId string, startDate, endDate time.Time) (*models.StatsSummary, error) {
	collection := s.getTimerSessionsCollection()

	// MongoDB aggregation pipeline to group by tag and sum durations
	pipeline := mongo.Pipeline{
		// Match user's completed sessions within the time range
		{{Key: "$match", Value: bson.M{
			"user_id": userId,
			"status":  models.StatusCompleted,
			"start_time": bson.M{
				"$gte": startDate,
				"$lte": endDate,
			},
		}}},
		// Group by tag and calculate totals
		{{Key: "$group", Value: bson.M{
			"_id":            "$tag",
			"total_duration": bson.M{"$sum": "$duration"},
			"session_count":  bson.M{"$sum": 1},
		}}},
		// Sort by total duration descending
		{{Key: "$sort", Value: bson.M{"total_duration": -1}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		if err = cursor.Close(ctx); err != nil {
			return
		}
	}(cursor, ctx)

	var tagStatsList []models.TagStats
	if err = cursor.All(ctx, &tagStatsList); err != nil {
		return nil, err
	}

	// Calculate summary statistics
	summary := &models.StatsSummary{
		TagBreakdown: tagStatsList,
	}

	for i := range tagStatsList {
		summary.TotalDuration += tagStatsList[i].TotalDuration
		summary.TotalSessions += tagStatsList[i].SessionCount
	}

	// Calculate percentages and averages for each tag
	for i := range tagStatsList {
		if tagStatsList[i].SessionCount > 0 {
			tagStatsList[i].AverageSession = tagStatsList[i].TotalDuration / int64(tagStatsList[i].SessionCount)
		}
		if summary.TotalDuration > 0 {
			tagStatsList[i].PercentageOfTotal = float64(tagStatsList[i].TotalDuration) / float64(summary.TotalDuration) * 100
		}
	}

	// Set most used tag (first one after sorting by duration desc)
	if len(tagStatsList) > 0 {
		summary.MostUsedTag = tagStatsList[0].Tag
	}

	// Calculate overall average session duration
	if summary.TotalSessions > 0 {
		summary.AverageSession = summary.TotalDuration / int64(summary.TotalSessions)
	}

	return summary, nil
}

// GetTagSessions retrieves individual timer sessions for a specific tag within a time period
func (s *service) GetTagSessions(ctx context.Context, userId, tag string, startDate, endDate time.Time) ([]*models.TimerSession, error) {
	collection := s.getTimerSessionsCollection()

	filter := bson.M{
		"user_id": userId,
		"tag":     tag,
		"status":  models.StatusCompleted,
		"start_time": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}

	// Sort by start_time descending (most recent first)
	cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.M{"start_time": -1}))
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		if err = cursor.Close(ctx); err != nil {
			return
		}
	}(cursor, ctx)

	var sessions []*models.TimerSession
	if err = cursor.All(ctx, &sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (s *service) DeleteTimerSession(ctx context.Context, userId, tag string) error {
	collection := s.getTimerSessionsCollection()

	filter := bson.M{"user_id": userId, "tag": tag}

	if _, err := collection.DeleteMany(ctx, filter); err != nil {
		return err
	}
	return nil
}
