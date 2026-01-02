package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/neilsmahajan/productivity-timer/internal/models"
)

type Service interface {
	Health() map[string]string
	FindOrCreateUser(ctx context.Context, user *models.User) (*models.User, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	UpdateTimerSession(ctx context.Context, timerSession *models.TimerSession) error
	CreateTimerSession(ctx context.Context, timerSession *models.TimerSession) error
	FindTimerSession(ctx context.Context, userId, tag string, status models.TimerStatus) (*models.TimerSession, error)
	AbandonRunningTimers(ctx context.Context, userId, tag string) error
	UpdateUserTagStats(ctx context.Context, userTagStats *models.UserTagStats) error
	CreateUserTagStats(ctx context.Context, userTagStats *models.UserTagStats) error
	FindUserTagStats(ctx context.Context, userId string, tag string) (*models.UserTagStats, error)
	FindAllUserTagStats(ctx context.Context, userId string) ([]*models.UserTagStats, error)
	GetStatsSummary(ctx context.Context, userId string, startDate, endDate time.Time) (*models.StatsSummary, error)
	GetTagSessions(ctx context.Context, userId, tag string, startDate, endDate time.Time) ([]*models.TimerSession, error)
	DeleteUserTagStats(ctx context.Context, userId, tag string) error
	DeleteTimerSession(ctx context.Context, userId, tag string) error
}

type service struct {
	db *mongo.Client
}

var (
	// MongoDB Atlas connection string (preferred for production)
	mongoURI = os.Getenv("MONGODB_URI")
	// Legacy environment variables for local development
	host     = os.Getenv("DB_HOST")
	port     = os.Getenv("DB_PORT")
	database = os.Getenv("DB_DATABASE")
	username = os.Getenv("DB_USERNAME")
	password = os.Getenv("DB_ROOT_PASSWORD")
)

func New() Service {
	var uri string

	// Use MONGODB_URI if provided (Atlas), otherwise construct from parts (local)
	if mongoURI != "" {
		uri = mongoURI
	} else if username != "" && password != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%s", username, password, host, port)
	} else {
		uri = fmt.Sprintf("mongodb://%s:%s", host, port)
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	return &service{
		db: client,
	}
}

func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := s.db.Ping(ctx, nil)
	if err != nil {
		return map[string]string{
			"status":  "unhealthy",
			"message": "Database connection failed",
		}
	}

	return map[string]string{
		"status":  "healthy",
		"message": "It's healthy",
	}
}
